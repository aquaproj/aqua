package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func (inst *installer) download(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest, assetName string, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})

	if pkgInfo.Type == "go_install" {
		return inst.downloadGoInstall(ctx, pkg, pkgInfo, dest, logE)
	}

	logE.Info("download and unarchive the package")

	body, err := inst.packageDownloader.GetReadCloser(ctx, pkg, pkgInfo, assetName, logE)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	return unarchive.Unarchive(&unarchive.File{ //nolint:wrapcheck
		Body:     body,
		Filename: assetName,
		Type:     pkgInfo.GetFormat(),
	}, dest, logE, inst.fs)
}

func (inst *installer) downloadGoInstall(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest string, logE *logrus.Entry) error {
	goPkgPath := pkgInfo.GetPath() + "@" + pkg.Version
	logE.WithFields(logrus.Fields{
		"gobin":           dest,
		"go_package_path": goPkgPath,
	}).Info("Installing a Go tool")
	if _, err := inst.executor.GoInstall(ctx, goPkgPath, dest); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}
