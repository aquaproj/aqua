package registry

import "github.com/aquaproj/aqua/v2/pkg/runtime"

func (pkgInfo *PackageInfo) Override(v string, rt *runtime.Runtime) (*PackageInfo, error) {
	pkg, err := pkgInfo.SetVersion(v)
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
