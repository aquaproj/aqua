package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/sirupsen/logrus"
)

func (is *Installer) verifyWithGitHubArtifactAttestation(ctx context.Context, logE *logrus.Entry, pkg *config.Package, gaa *registry.GitHubArtifactAttestations, bodyFile *download.DownloadedFile) error {
	if !gaa.GetEnabled() {
		return nil
	}

	tempFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}

	logE.Info("verify GitHub Artifact Attestations")
	if err := is.ghInstaller.install(ctx, logE); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}

	if err := is.ghVerifier.Verify(ctx, logE, &ghattestation.ParamVerify{
		Repository:     pkg.PackageInfo.RepoOwner + "/" + pkg.PackageInfo.RepoName,
		ArtifactPath:   tempFilePath,
		SignerWorkflow: gaa.SignerWorkflow,
	}); err != nil {
		return fmt.Errorf("verify a package with minisign: %w", err)
	}

	return nil
}
