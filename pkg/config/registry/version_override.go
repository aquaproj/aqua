package registry

import (
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (pkgInfo *PackageInfo) setTopVersion(logE *logrus.Entry, v string) *PackageInfo {
	sv := v
	if pkgInfo.VersionPrefix != nil {
		prefix := *pkgInfo.VersionPrefix
		if !strings.HasPrefix(v, prefix) {
			return nil
		}
		sv = strings.TrimPrefix(v, prefix)
	}
	a, err := expr.EvaluateVersionConstraints(pkgInfo.VersionConstraints, v, sv)
	if err != nil {
		// If it fails to evaluate version_constraint, output a debug log and treats as version_constraint is false.
		logerr.WithError(logE, err).Debug("evaluate the version_constraint")
		return nil
	}
	if a {
		return pkgInfo.Copy()
	}
	return nil
}

func (pkgInfo *PackageInfo) SetVersion(logE *logrus.Entry, v string) (*PackageInfo, error) {
	if pkgInfo.VersionConstraints == "" {
		return pkgInfo, nil
	}

	if p := pkgInfo.setTopVersion(logE, v); p != nil {
		return p, nil
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
			// If it fails to evaluate version_constraint, output a debug log and treats as version_constraint is false.
			logerr.WithError(logE, err).Debug("evaluate the version_constraint")
			continue
		}
		if a {
			return pkgInfo.overrideVersion(vo), nil
		}
	}
	return pkgInfo.Copy(), nil
}
