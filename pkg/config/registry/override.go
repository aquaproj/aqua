package registry

import (
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func (p *PackageInfo) Override(logger *slog.Logger, v string, rt *runtime.Runtime) (*PackageInfo, error) {
	pkg, err := p.SetVersion(logger, v)
	if err != nil {
		return nil, err
	}
	pkg.OverrideByRuntime(rt)
	return pkg, nil
}

func (p *PackageInfo) getOverride(rt *runtime.Runtime) *Override {
	for _, ov := range p.Overrides {
		if ov.Match(rt) {
			return ov
		}
	}
	return nil
}

func (ov *Override) Match(rt *runtime.Runtime) bool {
	if ov.GOOS != "" && ov.GOOS != rt.GOOS {
		return false
	}
	if ov.GOArch != "" && ov.GOArch != rt.GOARCH {
		return false
	}
	if ov.Envs != nil {
		if !matchEnvs(ov.Envs, rt.GOOS, rt.GOARCH, rt.GOOS+"/"+rt.GOARCH, false, false) {
			return false
		}
	}
	return true
}
