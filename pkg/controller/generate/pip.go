package generate

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) getPipVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *FindingPackage) string {
	pkgInfo := pkg.PackageInfo
	pipName := *pkgInfo.PipName
	if param.SelectVersion {
		versionStrings, err := ctrl.pipClient.ListVersions(ctx, pipName)
		if err != nil {
			logE.WithError(err).Warn("list versions")
			return ""
		}
		versions := make([]*Version, len(versionStrings))
		for i, v := range versionStrings {
			versions[i] = &Version{
				Version: v,
				URL:     fmt.Sprintf("https://pypi.org/project/%s/%s/", pipName, v),
			}
		}
		idx, err := ctrl.versionSelector.Find(versions)
		if err != nil {
			return ""
		}
		return versions[idx].Version
	}
	version, err := ctrl.pipClient.GetLatestVersion(ctx, pipName)
	if err != nil {
		logE.WithError(err).Warn("get a latest version")
		return ""
	}
	return version
}
