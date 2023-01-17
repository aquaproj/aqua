package registry

import (
	"github.com/aquaproj/aqua/pkg/runtime"
)

func (pkgInfo *PackageInfo) CheckSupported(rt *runtime.Runtime, env string) (bool, error) {
	if pkgInfo.SupportedEnvs != nil {
		return pkgInfo.CheckSupportedEnvs(rt.GOOS, rt.GOARCH, env), nil
	}
	return true, nil
}

func (pkgInfo *PackageInfo) CheckSupportedEnvs(goos, goarch, env string) bool {
	if pkgInfo.SupportedEnvs == nil {
		return true
	}
	if goos == "darwin" && goarch == "arm64" && pkgInfo.GetRosetta2() {
		return true
	}
	for _, supportedEnv := range pkgInfo.SupportedEnvs {
		switch supportedEnv {
		case goos, goarch, env, "all":
			return true
		}
	}
	return false
}
