package controller

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) download(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest, assetName string) error {
	logE := ctrl.logE().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Info("download and unarchive the package")

	body, err := ctrl.PackageDownloader.GetReadCloser(ctx, pkg, pkgInfo, assetName)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err //nolint:wrapcheck
	}

	return unarchive.Unarchive(body, assetName, pkgInfo.GetFormat(), dest) //nolint:wrapcheck
}
