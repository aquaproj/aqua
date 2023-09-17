package generate

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
)

func (c *Controller) getCargoVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *fuzzyfinder.Package) string {
	pkgInfo := pkg.PackageInfo
	if param.SelectVersion {
		versionStrings, err := c.cargoClient.ListVersions(ctx, pkgInfo.Crate)
		if err != nil {
			logE.WithError(err).Warn("list versions")
			return ""
		}
		versions := fuzzyfinder.ConvertStringsToVersions(versionStrings)

		idx, err := c.versionSelector.Find(versions, false)
		if err != nil {
			return ""
		}
		return versions[idx].Version
	}
	version, err := c.cargoClient.GetLatestVersion(ctx, pkgInfo.Crate)
	if err != nil {
		logE.WithError(err).Warn("get a latest version")
		return ""
	}
	return version
}
