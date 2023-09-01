package registry

import (
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (p *PackageInfo) setTopVersion(logE *logrus.Entry, v string) *PackageInfo {
	sv := v
	if p.VersionPrefix != "" {
		prefix := p.VersionPrefix
		if !strings.HasPrefix(v, prefix) {
			return nil
		}
		sv = strings.TrimPrefix(v, prefix)
	}
	a, err := expr.EvaluateVersionConstraints(p.VersionConstraints, v, sv)
	if err != nil {
		// If it fails to evaluate version_constraint, output a debug log and treats as version_constraint is false.
		logerr.WithError(logE, err).Debug("evaluate the version_constraint")
		return nil
	}
	if a {
		logE.WithFields(logrus.Fields{
			"version_constraint": p.VersionConstraints,
			"package_version":    v,
			"package_semver":     sv,
		}).Debug("match the version_constraint")
		return p.Copy()
	}
	return nil
}

func (p *PackageInfo) SetVersion(logE *logrus.Entry, v string) (*PackageInfo, error) {
	if p.VersionConstraints == "" {
		logE.Debug("no version_constraint")
		return p, nil
	}

	if p2 := p.setTopVersion(logE, v); p2 != nil {
		return p2, nil
	}

	for _, vo := range p.VersionOverrides {
		sv := v
		vp := p.VersionPrefix
		if vo.VersionPrefix != nil {
			vp = *vo.VersionPrefix
		}
		if vp != "" {
			prefix := vp
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
			logE.WithFields(logrus.Fields{
				"version_constraint": vo.VersionConstraints,
				"package_version":    v,
				"package_semver":     sv,
			}).Debug("match the version_constraint")
			return p.overrideVersion(vo), nil
		}
	}
	logE.WithFields(logrus.Fields{
		"version_constraint": p.VersionConstraints,
		"package_version":    v,
	}).Debug("no version_constraint matches")
	return p.Copy(), nil
}
