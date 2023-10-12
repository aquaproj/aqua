package generate

// import (
// 	"context"
//
// 	"github.com/aquaproj/aqua/v2/pkg/config"
// 	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
// 	"github.com/sirupsen/logrus"
// )
//
// func (c *Controller) getCargoVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *fuzzyfinder.Package) string {
// 	pkgInfo := pkg.PackageInfo
// 	if param.SelectVersion {
// 		versionStrings, err := c.cargoClient.ListVersions(ctx, pkgInfo.Crate)
// 		if err != nil {
// 			logE.WithError(err).Warn("list versions")
// 			return ""
// 		}
// 		items := fuzzyfinder.ConvertStringsToItems(versionStrings)
//
// 		idx, err := c.fuzzyFinder.Find(items, false)
// 		if err != nil {
// 			return ""
// 		}
// 		return items[idx].Item
// 	}
// 	version, err := c.cargoClient.GetLatestVersion(ctx, pkgInfo.Crate)
// 	if err != nil {
// 		logE.WithError(err).Warn("get a latest version")
// 		return ""
// 	}
// 	return version
// }
