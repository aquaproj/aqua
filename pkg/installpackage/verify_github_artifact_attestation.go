package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/sirupsen/logrus"
)

func (is *Installer) verifyWithGitHubArtifactAttestation(ctx context.Context, logE *logrus.Entry, bodyFile *download.DownloadedFile, param *DownloadParam) error {
	ppkg := param.Package
	m := ppkg.PackageInfo.GitHubArtifactAttestations
	if !m.GetEnabled() {
		return nil
	}
	logE.Info("verify a package with GitHub Artifact Attestations")
	if err := is.ghInstaller.install(ctx, logE); err != nil {
		return fmt.Errorf("install GitHub CLI: %w", err)
	}
	tempFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	if err := is.ghVerifier.Verify(ctx, logE, &ghattestation.ParamVerify{
		ArtifactPath:   tempFilePath,
		Repository:     ppkg.PackageInfo.RepoOwner + "/" + ppkg.PackageInfo.RepoName,
		SignerWorkflow: m.SignerWorkflow,
	}); err != nil {
		return fmt.Errorf("verify a package with minisign: %w", err)
	}
	return nil
}
