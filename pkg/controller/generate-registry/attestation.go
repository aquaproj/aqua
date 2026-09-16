package genrgst

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
)

const (
	predicateTypeSLSAProvenanceV1 = "https://slsa.dev/provenance/v1"
	githubURLPrefix               = "https://github.com/"
)

// sigstoreBundle is the subset of a Sigstore bundle needed to read the in-toto statement.
// https://github.com/sigstore/protobuf-specs/blob/main/protos/sigstore_bundle.proto
type sigstoreBundle struct {
	DSSEEnvelope *dsseEnvelope `json:"dsseEnvelope"`
}

type dsseEnvelope struct {
	Payload string `json:"payload"`
}

// provenanceStatement is the subset of a SLSA provenance v1 in-toto statement
// needed to determine the signer workflow.
// https://slsa.dev/spec/v1.0/provenance
type provenanceStatement struct {
	PredicateType string              `json:"predicateType"`
	Predicate     provenancePredicate `json:"predicate"`
}

type provenancePredicate struct {
	BuildDefinition struct {
		ExternalParameters struct {
			Workflow struct {
				Path       string `json:"path"`
				Repository string `json:"repository"`
			} `json:"workflow"`
		} `json:"externalParameters"`
	} `json:"buildDefinition"`
}

// patchGitHubArtifactAttestations checks whether the release publishes GitHub artifact
// attestations (build provenance) and, if so, sets github_artifact_attestations so that
// aqua verifies the attestation on install.
//
// It probes at most two assets of a single release (one tool asset and one checksum file),
// so the number of extra API calls per package stays small even though `aqua gr` checks
// all versions. Attestations are only looked up for the release the caller passes
// (the latest one); older versions typically predate the attestation setup.
func (c *Controller) patchGitHubArtifactAttestations(ctx context.Context, logger *slog.Logger, pkgInfo *registry.PackageInfo, tagName string, assets []*github.ReleaseAsset) {
	toolAsset, checksumAsset := selectAttestationSubjects(tagName, assets)
	// Prefer the attestation of the tool asset itself.
	if sw := c.getSignerWorkflow(ctx, logger, pkgInfo, toolAsset); sw != "" {
		pkgInfo.GitHubArtifactAttestations = &registry.GitHubArtifactAttestations{
			SignerWorkflow2: sw,
		}
		return
	}
	// Some projects attest only the checksum file (e.g. terraform-linters/tflint).
	// Then the tool asset is verified transitively via its checksum.
	if pkgInfo.Checksum == nil {
		return
	}
	if sw := c.getSignerWorkflow(ctx, logger, pkgInfo, checksumAsset); sw != "" {
		pkgInfo.Checksum.GitHubArtifactAttestations = &registry.GitHubArtifactAttestations{
			SignerWorkflow2: sw,
		}
	}
}

// isNonToolAsset returns true for assets which are signatures, certificates, or metadata
// rather than the tool itself. They aren't used as attestation subjects.
func isNonToolAsset(assetName string) bool {
	for _, suffix := range []string{
		".asc",
		".bundle",
		".intoto.jsonl",
		".pem",
		".pub",
		".sbom",
		".sbom.json",
		".sig",
		".sigstore",
		".sigstore.json",
		".spdx.json",
	} {
		if strings.HasSuffix(assetName, suffix) {
			return true
		}
	}
	return false
}

// selectAttestationSubjects picks the assets whose attestations are looked up:
// the first tool asset and the first checksum file, sorted by name for determinism.
// Assets without a digest are skipped because the attestation API requires one.
func selectAttestationSubjects(tagName string, assets []*github.ReleaseAsset) (*github.ReleaseAsset, *github.ReleaseAsset) {
	sorted := make([]*github.ReleaseAsset, len(assets))
	copy(sorted, assets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetName() < sorted[j].GetName()
	})
	var toolAsset, checksumAsset *github.ReleaseAsset
	for _, a := range sorted {
		if a.GetDigest() == "" {
			continue
		}
		name := a.GetName()
		if checksum.GetChecksumConfigFromFilename(name, tagName) != nil {
			if checksumAsset == nil {
				checksumAsset = a
			}
			continue
		}
		if isNonToolAsset(name) {
			continue
		}
		if toolAsset == nil {
			toolAsset = a
		}
		if toolAsset != nil && checksumAsset != nil {
			break
		}
	}
	return toolAsset, checksumAsset
}

// getSignerWorkflow returns the signer workflow of the asset's build provenance
// attestation, or an empty string if the asset has none.
func (c *Controller) getSignerWorkflow(ctx context.Context, logger *slog.Logger, pkgInfo *registry.PackageInfo, asset *github.ReleaseAsset) string {
	if asset == nil {
		return ""
	}
	attestations, _, err := c.github.ListAttestations(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, asset.GetDigest(), nil)
	if err != nil {
		// A missing attestation isn't an error of `aqua gr`; the package simply
		// doesn't get a github_artifact_attestations configuration.
		logger.Debug("list attestations",
			"repo_owner", pkgInfo.RepoOwner,
			"repo_name", pkgInfo.RepoName,
			"asset_name", asset.GetName(),
			"error", err,
		)
		return ""
	}
	if attestations == nil {
		return ""
	}
	for _, attestation := range attestations.Attestations {
		if sw := signerWorkflowFromBundle(attestation.Bundle); sw != "" {
			return sw
		}
	}
	return ""
}

// signerWorkflowFromBundle extracts the signer workflow (owner/repo/.github/workflows/x.yml)
// from a Sigstore bundle of a SLSA provenance v1 attestation.
//
// Attestations with other predicate types are ignored. Especially, GitHub's release
// attestations (https://in-toto.io/attestation/release/v0.2), which immutable releases
// generate automatically, must not be picked up: aqua intentionally doesn't verify them.
// https://github.com/aquaproj/aqua/issues/4862
//
// The signer workflow is taken from the provenance payload (externalParameters.workflow).
// When a release is signed via a reusable workflow, the certificate identity can differ
// from this value; in that case the generated configuration fails verification and has
// to be fixed manually, which is still detected before the registry change is merged
// because `cmdx t` verifies the package.
func signerWorkflowFromBundle(bundle []byte) string {
	b := &sigstoreBundle{}
	if err := json.Unmarshal(bundle, b); err != nil {
		return ""
	}
	if b.DSSEEnvelope == nil {
		return ""
	}
	payload, err := base64.StdEncoding.DecodeString(b.DSSEEnvelope.Payload)
	if err != nil {
		return ""
	}
	statement := &provenanceStatement{}
	if err := json.Unmarshal(payload, statement); err != nil {
		return ""
	}
	if statement.PredicateType != predicateTypeSLSAProvenanceV1 {
		return ""
	}
	workflow := statement.Predicate.BuildDefinition.ExternalParameters.Workflow
	repo, ok := strings.CutPrefix(workflow.Repository, githubURLPrefix)
	if !ok || repo == "" || workflow.Path == "" {
		return ""
	}
	return repo + "/" + workflow.Path
}
