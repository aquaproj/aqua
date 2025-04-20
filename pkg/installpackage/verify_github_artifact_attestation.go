package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/sirupsen/logrus"
)

type FileVerifier interface {
	Enabled(logE *logrus.Entry) (bool, error)
	Verify(ctx context.Context, logE *logrus.Entry, file string) error
}

type gitHubArtifactAttestationsVerifier struct {
	disabled    bool
	gaa         *registry.GitHubArtifactAttestations
	pkg         *config.Package
	ghInstaller *DedicatedInstaller
	ghVerifier  GitHubArtifactAttestationsVerifier
}

func (g *gitHubArtifactAttestationsVerifier) Enabled(logE *logrus.Entry) (bool, error) {
	if g.disabled {
		logE.Debug("GitHub Artifact Attestation is disabled")
		return false, nil
	}
	return g.gaa.GetEnabled(), nil
}

func (g *gitHubArtifactAttestationsVerifier) Verify(ctx context.Context, logE *logrus.Entry, file string) error {
	logE.Info("verify GitHub Artifact Attestations")
	if err := g.ghInstaller.install(ctx, logE); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}

	if err := g.ghVerifier.Verify(ctx, logE, &ghattestation.ParamVerify{
		Repository:     g.pkg.PackageInfo.RepoOwner + "/" + g.pkg.PackageInfo.RepoName,
		ArtifactPath:   file,
		PredicateType:  g.gaa.PredicateType,
		SignerWorkflow: g.gaa.SignerWorkflow(),
	}); err != nil {
		return fmt.Errorf("verify a package with gh attestation: %w", err)
	}
	return nil
}
