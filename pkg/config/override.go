package config

import "github.com/aquaproj/aqua/pkg/runtime"

func (pkgInfo *PackageInfo) Override(v string, rt *runtime.Runtime) error {
	if err := pkgInfo.setVersion(v); err != nil {
		return err
	}
	pkgInfo.override(rt)
	return nil
}

func (pkgInfo *PackageInfo) getOverride(rt *runtime.Runtime) *Override {
	for _, ov := range pkgInfo.Overrides {
		if ov.Match(rt) {
			return ov
		}
	}
	return nil
}

func (pkgInfo *PackageInfo) override(rt *runtime.Runtime) {
	for _, fo := range pkgInfo.FormatOverrides {
		if fo.GOOS == rt.GOOS {
			pkgInfo.Format = fo.Format
			break
		}
	}

	ov := pkgInfo.getOverride(rt)
	if ov == nil {
		return
	}

	if pkgInfo.Replacements == nil {
		pkgInfo.Replacements = ov.Replacements
	} else {
		for k, v := range ov.Replacements {
			pkgInfo.Replacements[k] = v
		}
	}

	if ov.Format != "" {
		pkgInfo.Format = ov.Format
	}

	if ov.Asset != nil {
		pkgInfo.Asset = ov.Asset
	}

	if ov.Files != nil {
		pkgInfo.Files = ov.Files
	}

	if ov.URL != nil {
		pkgInfo.URL = ov.URL
	}
}
