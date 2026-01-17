package installpackage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
)

type gitHubReleaseAttestationsVerifier struct {
	disabled    bool
	gra         bool
	pkg         *config.Package
	ghInstaller *DedicatedInstaller
	ghVerifier  GitHubArtifactAttestationsVerifier
}

func (g *gitHubReleaseAttestationsVerifier) Enabled(logger *slog.Logger) (bool, error) {
	if g.disabled {
		logger.Debug("GitHub Release Attestation is disabled")
		return false, nil
	}
	return g.gra, nil
}

func (g *gitHubReleaseAttestationsVerifier) Verify(ctx context.Context, logger *slog.Logger, file string) error {
	logger.Info("verify GitHub Release Attestations")
	if err := g.ghInstaller.install(ctx, logger); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}

	if err := g.ghVerifier.VerifyRelease(ctx, logger, &ghattestation.ParamVerifyRelease{
		Repository:   g.pkg.PackageInfo.RepoOwner + "/" + g.pkg.PackageInfo.RepoName,
		ArtifactPath: file,
		Version:      g.pkg.Package.Version,
	}); err != nil {
		return fmt.Errorf("verify a package with gh attestation: %w", err)
	}
	return nil
}
