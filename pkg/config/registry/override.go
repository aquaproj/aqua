package registry

import (
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// supportedVariantKeys is the set of variant keys aqua knows how to evaluate.
// An override that references a key not listed here is treated as non-matching,
// which lets registry authors stack a more specific entry alongside a fallback
// override (without variants) for older aqua versions that do not yet recognize
// the key.
var supportedVariantKeys = map[string]struct{}{ //nolint:gochecknoglobals
	"libc": {},
}

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
	for _, v := range ov.Variants {
		if _, ok := supportedVariantKeys[v.Key]; !ok {
			return false
		}
		if runtimeVariantValue(rt, v.Key) != v.Value {
			return false
		}
	}
	return true
}

func runtimeVariantValue(rt *runtime.Runtime, key string) string {
	if key == "libc" {
		return rt.LibC
	}
	return ""
}
