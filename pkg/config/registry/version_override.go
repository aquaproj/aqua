package registry

import "github.com/aquaproj/aqua/pkg/expr"

func (pkgInfo *PackageInfo) SetVersion(v string) (*PackageInfo, error) {
	if pkgInfo.VersionConstraints == "" {
		return pkgInfo, nil
	}
	a, err := expr.EvaluateVersionConstraints(pkgInfo.VersionConstraints, v)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if a {
		return pkgInfo.Copy(), nil
	}
	for _, vo := range pkgInfo.VersionOverrides {
		a, err := expr.EvaluateVersionConstraints(vo.VersionConstraints, v)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		if a {
			return pkgInfo.overrideVersion(vo), nil
		}
	}
	return nil, errNoVersionConstraintMatch
}
