package registry

import (
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
)

func (pkgInfo *PackageInfo) Override(logE *logrus.Entry, v string, rt *runtime.Runtime) (*PackageInfo, error) {
	pkg, err := pkgInfo.SetVersion(logE, v)
	if err != nil {
		return nil, err
	}
	pkg.OverrideByRuntime(rt)
	return pkg, nil
}

func (pkgInfo *PackageInfo) getOverride(rt *runtime.Runtime) *Override {
	for _, ov := range pkgInfo.Overrides {
		if ov.Match(rt) {
			return ov
		}
	}
	return nil
}
