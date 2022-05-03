package config

import constraint "github.com/aquaproj/aqua/pkg/version-constraint"

func (pkgInfo *PackageInfo) setVersion(v string) (*PackageInfo, error) {
	if pkgInfo.VersionConstraints == "" {
		return pkgInfo, nil
	}
	a, err := constraint.EvaluateVersionConstraints(pkgInfo.VersionConstraints, v)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if a {
		return pkgInfo.copy(), nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := constraint.EvaluateVersionConstraints(vo.VersionConstraints, v)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		if a {
			return pkgInfo.overrideVersion(vo), nil
		}
	}
	return pkgInfo.copy(), nil
}
