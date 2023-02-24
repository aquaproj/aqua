package registry

import (
	"strings"

	"github.com/aquaproj/aqua/pkg/expr"
)

func (pkgInfo *PackageInfo) setTopVersion(v string) (*PackageInfo, error) {
	sv := v
	if pkgInfo.VersionPrefix != nil {
		prefix := *pkgInfo.VersionPrefix
		if !strings.HasPrefix(v, prefix) {
			return nil, nil //nolint:nilnil
		}
		sv = strings.TrimPrefix(v, prefix)
	}
	a, err := expr.EvaluateVersionConstraints(pkgInfo.VersionConstraints, v, sv)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if a {
		return pkgInfo.Copy(), nil
	}
	return nil, nil //nolint:nilnil
}

func (pkgInfo *PackageInfo) SetVersion(v string) (*PackageInfo, error) {
	if pkgInfo.VersionConstraints == "" {
		return pkgInfo, nil
	}
	p, err := pkgInfo.setTopVersion(v)
	if err != nil || p != nil {
		return p, err
	}
	for _, vo := range pkgInfo.VersionOverrides {
		sv := v
		p := pkgInfo.VersionPrefix
		if vo.VersionPrefix != nil {
			p = vo.VersionPrefix
		}
		if p != nil {
			prefix := *p
			if !strings.HasPrefix(v, prefix) {
				continue
			}
			sv = strings.TrimPrefix(v, prefix)
		}
		a, err := expr.EvaluateVersionConstraints(vo.VersionConstraints, v, sv)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		if a {
			return pkgInfo.overrideVersion(vo), nil
		}
	}
	return nil, errNoVersionConstraintMatch
}
