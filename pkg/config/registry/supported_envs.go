package registry

import (
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func (p *PackageInfo) CheckSupported(rt *runtime.Runtime, env string) (bool, error) {
	if p.CheckSupportedEnvs(rt.GOOS, rt.GOARCH, env) {
		return true, nil
	}
	if !p.Build.CheckEnabled() {
		return false, nil
	}
	if p.checkExcludedEnvs(rt.GOOS, rt.GOARCH, env) {
		p.OverrideByBuild()
		return true, nil
	}
	return false, nil
}

func (p *PackageInfo) CheckSupportedEnvs(goos, goarch, env string) bool {
	if p.SupportedEnvs == nil {
		return true
	}
	return matchEnvs(p.SupportedEnvs, goos, goarch, env, p.Rosetta2)
}

func (p *PackageInfo) checkExcludedEnvs(goos, goarch, env string) bool {
	if p.Build.ExcludedEnvs == nil {
		return true
	}
	return !matchEnvs(p.Build.ExcludedEnvs, goos, goarch, env, p.Rosetta2)
}

func matchEnvs(envs []string, goos, goarch, env string, rosetta2 bool) bool {
	for _, elem := range envs {
		switch elem {
		case goos, goarch, env, "all":
			return true
		}
	}
	if goos == "darwin" && goarch == "arm64" && rosetta2 {
		for _, elem := range envs {
			switch elem {
			case "amd64", "darwin/amd64":
				return true
			}
		}
	}
	return false
}
