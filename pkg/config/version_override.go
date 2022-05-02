package config

func (pkgInfo *PackageInfo) setVersion(v string) (*PackageInfo, error) {
	if pkgInfo.VersionConstraints == nil {
		return pkgInfo, nil
	}
	a, err := pkgInfo.VersionConstraints.Check(v)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if a {
		return pkgInfo, nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := vo.VersionConstraints.Check(v)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		if a {
			return pkgInfo.overrideVersion(vo), nil
		}
	}
	return pkgInfo, nil
}
