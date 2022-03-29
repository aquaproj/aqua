package config

func (pkgInfo *PackageInfo) setVersion(v string) error {
	if pkgInfo.VersionConstraints == nil {
		return nil
	}
	a, err := pkgInfo.VersionConstraints.Check(v)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if a {
		return nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := vo.VersionConstraints.Check(v)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if a {
			pkgInfo.overrideVersion(vo)
			return nil
		}
	}
	return nil
}

func (pkgInfo *PackageInfo) overrideVersion(child *PackageInfo) { //nolint:cyclop
	if child.Type != "" {
		pkgInfo.Type = child.Type
	}
	if child.RepoOwner != "" {
		pkgInfo.RepoOwner = child.RepoOwner
	}
	if child.RepoName != "" {
		pkgInfo.RepoName = child.RepoName
	}
	if child.Asset != nil {
		pkgInfo.Asset = child.Asset
	}
	if child.Path != nil {
		pkgInfo.Path = child.Path
	}
	if child.Format != "" {
		pkgInfo.Format = child.Format
	}
	if child.Files != nil {
		pkgInfo.Files = child.Files
	}
	if child.URL != nil {
		pkgInfo.URL = child.URL
	}
	if child.Replacements != nil {
		pkgInfo.Replacements = child.Replacements
	}
	if child.Overrides != nil {
		pkgInfo.Overrides = child.Overrides
	}
	if child.FormatOverrides != nil {
		pkgInfo.FormatOverrides = child.FormatOverrides
	}
	if child.SupportedIf != nil {
		pkgInfo.SupportedIf = child.SupportedIf
	}
	if child.Rosetta2 != nil {
		pkgInfo.Rosetta2 = child.Rosetta2
	}
	if child.VersionFilter != nil {
		pkgInfo.VersionFilter = child.VersionFilter
	}
}
