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
	if !ov.MatchPlatform(rt) {
		return false
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

// MatchPlatform reports whether ov matches rt by GOOS, GOArch, and Envs only,
// ignoring Variants. Variant-expansion code uses this to find candidate
// overrides for a given platform before enumerating variant values.
func (ov *Override) MatchPlatform(rt *runtime.Runtime) bool {
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

// SupportedVariantKeys returns the variant keys aqua knows how to evaluate.
// Callers that need to enumerate variant values (e.g. update-checksum) should
// use this list rather than hard-coding key names.
func SupportedVariantKeys() []string {
	keys := make([]string, 0, len(supportedVariantKeys))
	for k := range supportedVariantKeys {
		keys = append(keys, k)
	}
	return keys
}

// IsSupportedVariantKey reports whether key is in the supportedVariantKeys
// allowlist.
func IsSupportedVariantKey(key string) bool {
	_, ok := supportedVariantKeys[key]
	return ok
}

func runtimeVariantValue(rt *runtime.Runtime, key string) string {
	if key == "libc" {
		return rt.LibC
	}
	return ""
}
