package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/sirupsen/logrus"
)

func (is *InstallerImpl) verifyWithSLSA(ctx context.Context, logE *logrus.Entry, bodyFile *download.DownloadedFile, param *DownloadParam) error {
	ppkg := param.Package
	pkgInfo := param.Package.PackageInfo
	sp := ppkg.PackageInfo.SLSAProvenance
	if !sp.GetEnabled() {
		return nil
	}
	art := ppkg.GetTemplateArtifact(is.runtime, param.Asset)
	logE.Info("verify a package with slsa-verifier")
	if err := is.slsaVerifierInstaller.installSLSAVerifier(ctx, logE, slsa.Version); err != nil {
		return fmt.Errorf("install slsa-verifier: %w", err)
	}
	tempFilePath, err := bodyFile.GetPath()
	if err != nil {
		return fmt.Errorf("get a temporal file path: %w", err)
	}
	if err := is.slsaVerifier.Verify(ctx, logE, is.runtime, sp, art, &download.File{
		RepoOwner: ppkg.PackageInfo.RepoOwner,
		RepoName:  ppkg.PackageInfo.RepoName,
		Version:   ppkg.Package.Version,
	}, &slsa.ParamVerify{
		SourceURI:    pkgInfo.SLSASourceURI(),
		SourceTag:    ppkg.Package.Version,
		ArtifactPath: tempFilePath,
	}); err != nil {
		return fmt.Errorf("verify a package with slsa-verifier: %w", err)
	}
	return nil
}
