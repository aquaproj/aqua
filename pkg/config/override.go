package config

import "github.com/aquaproj/aqua/pkg/runtime"

func (pkgInfo *PackageInfo) Override(v string, rt *runtime.Runtime) (*PackageInfo, error) {
	pkg, err := pkgInfo.setVersion(v)
	if err != nil {
		return nil, err
	}
	pkg.override(rt)
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
