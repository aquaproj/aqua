package genrgst

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/google/go-cmp/cmp"
)

func newTestBundle(t *testing.T, predicateType, repository, path string) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"predicateType": predicateType,
		"predicate": map[string]any{
			"buildDefinition": map[string]any{
				"externalParameters": map[string]any{
					"workflow": map[string]any{
						"repository": repository,
						"path":       path,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := json.Marshal(map[string]any{
		"dsseEnvelope": map[string]any{
			"payload": base64.StdEncoding.EncodeToString(payload),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestSignerWorkflowFromBundle(t *testing.T) {
	t.Parallel()
	data := []struct {
		name   string
		bundle func(t *testing.T) []byte
		exp    string
	}{
		{
			name: "slsa provenance v1",
			bundle: func(t *testing.T) []byte {
				t.Helper()
				return newTestBundle(t, "https://slsa.dev/provenance/v1", "https://github.com/evilmartians/lefthook", ".github/workflows/release.yml")
			},
			exp: "evilmartians/lefthook/.github/workflows/release.yml",
		},
		{
			name: "release attestation is ignored",
			bundle: func(t *testing.T) []byte {
				t.Helper()
				// GitHub's immutable releases generate release attestations automatically.
				// aqua intentionally doesn't verify them. https://github.com/aquaproj/aqua/issues/4862
				return newTestBundle(t, "https://in-toto.io/attestation/release/v0.2", "", "")
			},
			exp: "",
		},
		{
			name: "repository outside github.com is ignored",
			bundle: func(t *testing.T) []byte {
				t.Helper()
				return newTestBundle(t, "https://slsa.dev/provenance/v1", "https://example.com/foo/bar", ".github/workflows/release.yml")
			},
			exp: "",
		},
		{
			name: "workflow path is missing",
			bundle: func(t *testing.T) []byte {
				t.Helper()
				return newTestBundle(t, "https://slsa.dev/provenance/v1", "https://github.com/evilmartians/lefthook", "")
			},
			exp: "",
		},
		{
			name: "broken bundle",
			bundle: func(t *testing.T) []byte {
				t.Helper()
				return []byte("broken")
			},
			exp: "",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if sw := signerWorkflowFromBundle(d.bundle(t)); sw != d.exp {
				t.Fatalf("wanted %q, got %q", d.exp, sw)
			}
		})
	}
}

func TestSelectAttestationSubjects(t *testing.T) {
	t.Parallel()
	assets := []*github.ReleaseAsset{
		{
			Name:   new("tflint_darwin_arm64.zip.sig"),
			Digest: new("sha256:sig"),
		},
		{
			Name: new("tflint_windows_amd64.zip"),
			// Assets without digests can't be attestation subjects.
		},
		{
			Name:   new("tflint_darwin_arm64.zip"),
			Digest: new("sha256:tool"),
		},
		{
			Name:   new("checksums.txt"),
			Digest: new("sha256:checksum"),
		},
	}
	toolAsset, checksumAsset := selectAttestationSubjects("v0.64.0", assets)
	if toolAsset.GetName() != "tflint_darwin_arm64.zip" {
		t.Fatalf("wanted tflint_darwin_arm64.zip, got %q", toolAsset.GetName())
	}
	if checksumAsset.GetName() != "checksums.txt" {
		t.Fatalf("wanted checksums.txt, got %q", checksumAsset.GetName())
	}
}

func TestController_patchGitHubArtifactAttestations(t *testing.T) { //nolint:funlen
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	data := []struct {
		name         string
		pkgInfo      *registry.PackageInfo
		assets       []*github.ReleaseAsset
		attestations map[string]*github.AttestationsResponse
		exp          *registry.PackageInfo
	}{
		{
			name: "tool asset has build provenance",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "evilmartians",
				RepoName:  "lefthook",
			},
			assets: []*github.ReleaseAsset{
				{
					Name:   new("lefthook_2.1.12_MacOS_arm64.gz"),
					Digest: new("sha256:tool"),
				},
			},
			attestations: map[string]*github.AttestationsResponse{
				"sha256:tool": attestationsResponse(t, "https://slsa.dev/provenance/v1", "https://github.com/evilmartians/lefthook", ".github/workflows/release.yml"),
			},
			exp: &registry.PackageInfo{
				RepoOwner: "evilmartians",
				RepoName:  "lefthook",
				GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{
					SignerWorkflow2: "evilmartians/lefthook/.github/workflows/release.yml",
				},
			},
		},
		{
			name: "only the checksum file has build provenance",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "terraform-linters",
				RepoName:  "tflint",
				Checksum: &registry.Checksum{
					Type:      "github_release",
					Asset:     "checksums.txt",
					Algorithm: "sha256",
				},
			},
			assets: []*github.ReleaseAsset{
				{
					Name:   new("tflint_darwin_arm64.zip"),
					Digest: new("sha256:tool"),
				},
				{
					Name:   new("checksums.txt"),
					Digest: new("sha256:checksum"),
				},
			},
			attestations: map[string]*github.AttestationsResponse{
				"sha256:tool":     attestationsResponse(t, "https://in-toto.io/attestation/release/v0.2", "", ""),
				"sha256:checksum": attestationsResponse(t, "https://slsa.dev/provenance/v1", "https://github.com/terraform-linters/tflint", ".github/workflows/release.yml"),
			},
			exp: &registry.PackageInfo{
				RepoOwner: "terraform-linters",
				RepoName:  "tflint",
				Checksum: &registry.Checksum{
					Type:      "github_release",
					Asset:     "checksums.txt",
					Algorithm: "sha256",
					GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{
						SignerWorkflow2: "terraform-linters/tflint/.github/workflows/release.yml",
					},
				},
			},
		},
		{
			name: "no build provenance",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "open-policy-agent",
				RepoName:  "conftest",
			},
			assets: []*github.ReleaseAsset{
				{
					Name:   new("conftest_0.69.0_Darwin_arm64.tar.gz"),
					Digest: new("sha256:tool"),
				},
			},
			attestations: map[string]*github.AttestationsResponse{},
			exp: &registry.PackageInfo{
				RepoOwner: "open-policy-agent",
				RepoName:  "conftest",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			gh := &github.MockRepositoriesService{
				Attestations: d.attestations,
			}
			ctrl := NewController(gh, nil, nil, &bytes.Buffer{})
			ctrl.patchGitHubArtifactAttestations(ctx, logger, d.pkgInfo, "v1.0.0", d.assets)
			if diff := cmp.Diff(d.exp, d.pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func attestationsResponse(t *testing.T, predicateType, repository, path string) *github.AttestationsResponse {
	t.Helper()
	return &github.AttestationsResponse{
		Attestations: []*github.Attestation{
			{
				Bundle: newTestBundle(t, predicateType, repository, path),
			},
		},
	}
}
