package config

import "runtime"

func (pkgInfo *PackageInfo) Override(v string) error {
	if err := pkgInfo.setVersion(v); err != nil {
		return err
	}
	pkgInfo.override()
	return nil
}

func (pkgInfo *PackageInfo) getOverride() *Override {
	for _, ov := range pkgInfo.Overrides {
		if ov.Match() {
			return ov
		}
	}
	return nil
}

func (pkgInfo *PackageInfo) override() {
	for _, fo := range pkgInfo.FormatOverrides {
		if fo.GOOS == runtime.GOOS {
			pkgInfo.Format = fo.Format
			break
		}
	}

	ov := pkgInfo.getOverride()
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
