package installpackage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/download"
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

func (inst *InstallerImpl) download(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error { //nolint:funlen,cyclop
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

	if err := inst.verifyWithCosign(ctx, logE, bodyFile, param); err != nil {
		return err
	}

	if err := inst.verifyWithSLSA(ctx, logE, bodyFile, param); err != nil {
		return err
	}

	if err := inst.verifyChecksumWrap(ctx, logE, param, bodyFile); err != nil {
		return err
	}

	return inst.unarchiver.Unarchive(ctx, logE, &unarchive.File{ //nolint:wrapcheck
		Body:     bodyFile,
		Filename: param.Asset,
		Type:     pkgInfo.GetFormat(),
	}, param.Dest)
}
