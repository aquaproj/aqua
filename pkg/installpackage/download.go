package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) download(ctx context.Context, logE *logrus.Entry, param *DownloadParam) error {
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

	logE.Info("download and unarchive the package")

	body, cl, err := inst.packageDownloader.GetReadCloser(ctx, ppkg, param.Asset, logE)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	var pOpts *unarchive.ProgressBarOpts
	if inst.progressBar {
		pOpts = &unarchive.ProgressBarOpts{
			ContentLength: cl,
			Description:   fmt.Sprintf("Downloading %s %s", pkg.Name, pkg.Version),
		}
	}

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     body,
		Filename: param.Asset,
		Type:     pkgInfo.GetFormat(),
	}, param.Dest, logE, inst.fs, pOpts)
}

func (inst *Installer) downloadGoInstall(ctx context.Context, pkg *config.Package, dest string, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	goPkgPath := pkgInfo.GetPath() + "@" + pkg.Package.Version
	logE.WithFields(logrus.Fields{
		"gobin":           dest,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if _, err := inst.executor.GoInstall(ctx, goPkgPath, dest); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}
