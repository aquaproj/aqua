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
		Repository:     g.pkg.PackageInfo.RepoOwner + "/" + g.pkg.PackageInfo.RepoName,
		ArtifactPath:   file,
		PredicateType:  g.gaa.PredicateType,
		SignerWorkflow: g.gaa.SignerWorkflow(),
	}); err != nil {
		return fmt.Errorf("verify a package with gh attestation: %w", err)
	}
	return nil
}
