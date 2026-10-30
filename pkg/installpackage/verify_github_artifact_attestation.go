package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
)

type FileVerifier interface {
	// Kind names the verification as the registry does, e.g. "cosign". It is what
	// aqua-checksums.json records as verified.
	Kind() string
	Enabled(logger *slog.Logger) (bool, error)
	Verify(ctx context.Context, logger *slog.Logger, file string) error
}

type gitHubArtifactAttestationsVerifier struct {
	disabled    bool
	gaa         *registry.GitHubArtifactAttestations
	pkg         *config.Package
	ghInstaller *DedicatedInstaller
	ghVerifier  GitHubArtifactAttestationsVerifier
}

func (*gitHubArtifactAttestationsVerifier) Kind() string {
	return "github_artifact_attestations"
}

func (g *gitHubArtifactAttestationsVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if g.disabled {
		logger.Debug("GitHub Artifact Attestation is disabled")
		return false, nil
	}
	return g.gaa.GetEnabled(), nil
}

func (g *gitHubArtifactAttestationsVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify GitHub Artifact Attestations")
	if err := g.ghInstaller.install(ctx, logger); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}

	if err := g.ghVerifier.Verify(ctx, logger, &ghattestation.ParamVerify{
		Repository:            g.repository(),
		ArtifactPath:          file,
		PredicateType:         g.gaa.PredicateType,
		SignerWorkflow:        g.gaa.SignerWorkflow(),
		SignerRepo:            g.gaa.SignerRepo,
		SourceRef:             g.gaa.SourceRef,
		SourceDigest:          g.gaa.SourceDigest,
		CertOIDCIssuer:        g.gaa.CertOIDCIssuer,
		DenySelfHostedRunners: g.gaa.DenySelfHostedRunners,
	}); err != nil {
		return fmt.Errorf("verify a package with gh attestation: %w", err)
	}
	return nil
}

// repository is the repository the attestation is looked up in: the one the configuration
// names, or the package's.
func (g *gitHubArtifactAttestationsVerifier) repository() string {
	if g.gaa.Repository != "" {
		return g.gaa.Repository
	}
	return g.pkg.PackageInfo.RepoOwner + "/" + g.pkg.PackageInfo.RepoName
}
