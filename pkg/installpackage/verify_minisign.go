package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/sirupsen/logrus"
)

func (is *Installer) verifyWithMinisign(ctx context.Context, logE *logrus.Entry, bodyFile *download.DownloadedFile, param *DownloadParam) error {
	ppkg := param.Package
	m := ppkg.PackageInfo.Minisign
	if !m.GetEnabled() {
		return nil
	}
	mPkg := minisign.Package()
	if f, err := mPkg.PackageInfo.CheckSupported(is.realRuntime, is.realRuntime.Env()); err != nil {
		return fmt.Errorf("check if minisign supports this environment: %w", err)
	} else if !f {
		logE.Warn("minisign doesn't support this environment")
		return nil
	}
	art := ppkg.TemplateArtifact(is.runtime, param.Asset)
	logE.Info("verify a package with minisign")
	if err := is.minisignInstaller.install(ctx, logE); err != nil {
		return fmt.Errorf("install minisign: %w", err)
	}
	tempFilePath, err := bodyFile.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	if err := is.minisignVerifier.Verify(ctx, logE, is.runtime, m, art, &download.File{
		RepoOwner: ppkg.PackageInfo.RepoOwner,
		RepoName:  ppkg.PackageInfo.RepoName,
		Version:   ppkg.Package.Version,
	}, &minisign.ParamVerify{
		ArtifactPath: tempFilePath,
		PublicKey:    m.PublicKey,
	}); err != nil {
		return fmt.Errorf("verify a package with minisign: %w", err)
	}
	return nil
}
