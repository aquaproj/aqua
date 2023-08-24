package registry

import (
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
)

func (p *PackageInfo) Override(logE *logrus.Entry, v string, rt *runtime.Runtime) (*PackageInfo, error) {
	pkg, err := p.SetVersion(logE, v)
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
	return true
}
