package registry

import (
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func (p *PackageInfo) CheckSupported(rt *runtime.Runtime, env string) (bool, error) {
	if p.SupportedEnvs != nil {
		return p.CheckSupportedEnvs(rt.GOOS, rt.GOARCH, env), nil
	}
	return true, nil
}

func (p *PackageInfo) CheckSupportedEnvs(goos, goarch, env string) bool {
	if p.SupportedEnvs == nil {
		return true
	}
	if goos == "darwin" && goarch == "arm64" && p.GetRosetta2() {
		return true
	}
	for _, supportedEnv := range p.SupportedEnvs {
		switch supportedEnv {
		case goos, goarch, env, "all":
			return true
		}
	}
	return false
}
