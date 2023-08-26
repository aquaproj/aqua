package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/sirupsen/logrus"
)

func (is *InstallerImpl) verifyWithCosign(ctx context.Context, logE *logrus.Entry, bodyFile *download.DownloadedFile, param *DownloadParam) error {
	ppkg := param.Package

	cos := ppkg.PackageInfo.Cosign
	if !cos.GetEnabled() {
		return nil
	}

	art := ppkg.TemplateArtifact(is.runtime, param.Asset)
	logE.Info("verify a package with Cosign")
	if err := is.cosignInstaller.installCosign(ctx, logE, cosign.Version); err != nil {
		return fmt.Errorf("install sigstore/cosign: %w", err)
	}
	tempFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporal file path: %w", err)
	}
	if err := is.cosign.Verify(ctx, logE, is.runtime, &download.File{
		RepoOwner: ppkg.PackageInfo.RepoOwner,
		RepoName:  ppkg.PackageInfo.RepoName,
		Version:   ppkg.Package.Version,
	}, cos, art, tempFilePath); err != nil {
		return fmt.Errorf("verify a package with Cosign: %w", err)
	}
	return nil
}
