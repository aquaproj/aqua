package generate

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) getCargoVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *FindingPackage) string {
	pkgInfo := pkg.PackageInfo
	if param.SelectVersion {
		versionStrings, err := ctrl.cargoClient.ListVersions(ctx, *pkgInfo.Crate)
		if err != nil {
			logE.WithError(err).Warn("list versions")
			return ""
		}

		versions := convertStringsToVersions(versionStrings)

		idx, err := ctrl.versionSelector.Find(versions, false)
		if err != nil {
			return ""
		}
		return versions[idx].Version
	}
	version, err := ctrl.cargoClient.GetLatestVersion(ctx, *pkgInfo.Crate)
	if err != nil {
		logE.WithError(err).Warn("get a latest version")
		return ""
	}
	return version
}
