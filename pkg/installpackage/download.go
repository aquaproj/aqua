package installpackage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *InstallerImpl) downloadWithRetry(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    param.Package.Package.Name,
		"package_version": param.Package.Package.Version,
		"registry":        param.Package.Package.Registry,
	})
	retryCount := 0
	for {
		logE.Debug("check if the package is already installed")
		finfo, err := inst.fs.Stat(param.Dest)
		if err != nil { //nolint:nestif
			// file doesn't exist
			if err := inst.download(ctx, logE, param); err != nil {
				if strings.Contains(err.Error(), "file already exists") {
					if retryCount >= maxRetryDownload {
						return err
					}
					retryCount++
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"retry_count": retryCount,
					}).Info("retry installing the package")
					continue
				}
				return err
			}
			return nil
		}
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", param.Dest)
		}
		return nil
	}
}

func (inst *InstallerImpl) download(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error { //nolint:funlen,cyclop,gocognit
	ppkg := param.Package
	pkg := ppkg.Package
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	pkgInfo := param.Package.PackageInfo

	if pkgInfo.Type == "go_install" {
		return inst.downloadGoInstall(ctx, ppkg, param.Dest, logE)
	}

	if pkgInfo.Type == "cargo" {
		return inst.downloadCargo(ctx, logE, ppkg, param.Dest)
	}

	logE.Info("download and unarchive the package")

	file, err := download.ConvertPackageToFile(ppkg, param.Asset, inst.runtime)
	if err != nil {
		return err //nolint:wrapcheck
	}
	body, cl, err := inst.downloader.GetReadCloser(ctx, logE, file)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	var pb *progressbar.ProgressBar
	if inst.progressBar && cl != 0 {
		pb = progressbar.DefaultBytes(
			cl,
			fmt.Sprintf("Downloading %s %s", pkg.Name, pkg.Version),
		)
	}
	bodyFile := download.NewDownloadedFile(inst.fs, body, pb)
	defer func() {
		if err := bodyFile.Remove(); err != nil {
			logE.WithError(err).Warn("remove a temporal file")
		}
	}()

	// Verify with Cosign
	if cos := ppkg.PackageInfo.Cosign; cos.GetEnabled() {
		art := ppkg.GetTemplateArtifact(inst.runtime, param.Asset)
		logE.Info("verify a package with Cosign")
		if err := inst.cosignInstaller.installCosign(ctx, logE, cosign.Version); err != nil {
			return fmt.Errorf("install sigstore/cosign: %w", err)
		}
		tempFilePath, err := bodyFile.GetPath()
		if err != nil {
			return fmt.Errorf("get a temporal file path: %w", err)
		}
		if err := inst.cosign.Verify(ctx, logE, inst.runtime, &download.File{
			RepoOwner: ppkg.PackageInfo.RepoOwner,
			RepoName:  ppkg.PackageInfo.RepoName,
			Version:   ppkg.Package.Version,
		}, cos, art, tempFilePath); err != nil {
			return fmt.Errorf("verify a package with Cosign: %w", err)
		}
	}

	// Verify with SLSA Provenance
	if sp := ppkg.PackageInfo.SLSAProvenance; sp.GetEnabled() {
		art := ppkg.GetTemplateArtifact(inst.runtime, param.Asset)
		logE.Info("verify a package with slsa-verifier")
		if err := inst.slsaVerifierInstaller.installSLSAVerifier(ctx, logE, slsa.Version); err != nil {
			return fmt.Errorf("install slsa-verifier: %w", err)
		}
		tempFilePath, err := bodyFile.GetPath()
		if err != nil {
			return fmt.Errorf("get a temporal file path: %w", err)
		}
		if err := inst.slsaVerifier.Verify(ctx, logE, inst.runtime, sp, art, &download.File{
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
	}

	if param.Checksum != nil || param.Checksums != nil { //nolint:nestif
		tempFilePath, err := bodyFile.GetPath()
		if err != nil {
			return fmt.Errorf("get a temporal file path: %w", err)
		}
		paramVerifyChecksum := &ParamVerifyChecksum{
			Checksum:        param.Checksum,
			Checksums:       param.Checksums,
			Pkg:             ppkg,
			AssetName:       param.Asset,
			TempFilePath:    tempFilePath,
			SkipSetChecksum: true,
		}

		if param.Checksum == nil {
			paramVerifyChecksum.SkipSetChecksum = false
			cid, err := ppkg.GetChecksumID(inst.runtime)
			if err != nil {
				return err //nolint:wrapcheck
			}
			paramVerifyChecksum.ChecksumID = cid
			// Even if SLSA Provenance is enabled checksum verification is run
			paramVerifyChecksum.Checksum = param.Checksums.Get(cid)
			if paramVerifyChecksum.Checksum == nil && param.RequireChecksum {
				return logerr.WithFields(errChecksumIsRequired, logrus.Fields{ //nolint:wrapcheck
					"doc": "https://aquaproj.github.io/docs/reference/codes/001",
				})
			}
		}

		if err := inst.verifyChecksum(ctx, logE, paramVerifyChecksum); err != nil {
			return err
		}
	}

	return inst.unarchiver.Unarchive(ctx, logE, &unarchive.File{ //nolint:wrapcheck
		Body:     bodyFile,
		Filename: param.Asset,
		Type:     pkgInfo.GetFormat(),
	}, param.Dest)
}

func (inst *InstallerImpl) downloadGoInstall(ctx context.Context, pkg *config.Package, dest string, logE *logrus.Entry) error {
	p, err := pkg.RenderPath()
	if err != nil {
		return fmt.Errorf("render Go Module Path: %w", err)
	}
	goPkgPath := p + "@" + pkg.Package.Version
	logE.WithFields(logrus.Fields{
		"gobin":           dest,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if err := inst.goInstallInstaller.Install(ctx, goPkgPath, dest); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}

func (inst *InstallerImpl) downloadCargo(ctx context.Context, logE *logrus.Entry, pkg *config.Package, root string) error {
	logE.Info("Installing a crate")
	crate := *pkg.PackageInfo.Crate
	version := pkg.Package.Version
	if err := inst.cargoPackageInstaller.Install(ctx, crate, version, root); err != nil {
		return fmt.Errorf("cargo install: %w", err)
	}
	return nil
}
