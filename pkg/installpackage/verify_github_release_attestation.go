package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/sirupsen/logrus"
)

type gitHubReleaseAttestationsVerifier struct {
	disabled    bool
	gra         *registry.GitHubReleaseAttestation
	pkg         *config.Package
	ghInstaller *DedicatedInstaller
	ghVerifier  GitHubArtifactAttestationsVerifier
}

func (g *gitHubReleaseAttestationsVerifier) Enabled(logE *logrus.Entry) (bool, error) {
	if g.disabled {
		logE.Debug("GitHub Release Attestation is disabled")
		return false, nil
	}
	return g.gra.GetEnabled(), nil
}

func (g *gitHubReleaseAttestationsVerifier) Verify(ctx context.Context, logE *logrus.Entry, file string) error {
	logE.Info("verify GitHub Release Attestations")
	if err := g.ghInstaller.install(ctx, logE); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}

	if err := g.ghVerifier.VerifyRelease(ctx, logE, &ghattestation.ParamVerifyRelease{
		Repository:   g.pkg.PackageInfo.RepoOwner + "/" + g.pkg.PackageInfo.RepoName,
		ArtifactPath: file,
	}); err != nil {
		return fmt.Errorf("verify a package with gh attestation: %w", err)
	}
	return nil
}
