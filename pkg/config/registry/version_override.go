package registry

import (
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (p *PackageInfo) setTopVersion(logger *slog.Logger, v string) *PackageInfo {
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
		slogerr.WithError(logger, err).Debug("evaluate the version_constraint")
		return nil
	}
	if a {
		logger.Debug("match the version_constraint",
			slog.String("version_constraint", p.VersionConstraints),
			slog.String("package_version", v),
			slog.String("package_semver", sv),
		)
		return p.Copy()
	}
	return nil
}

func (p *PackageInfo) SetVersion(logger *slog.Logger, v string) (*PackageInfo, error) {
	if p.VersionConstraints == "" {
		logger.Debug("no version_constraint")
		return p, nil
	}

	if p2 := p.setTopVersion(logger, v); p2 != nil {
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
			slogerr.WithError(logger, err).Debug("evaluate the version_constraint")
			continue
		}
		if a {
			logger.Debug("match the version_constraint",
				slog.String("version_constraint", vo.VersionConstraints),
				slog.String("package_version", v),
				slog.String("package_semver", sv),
			)
			return p.overrideVersion(vo), nil
		}
	}
	logger.Debug("no version_constraint matches",
		slog.String("version_constraint", p.VersionConstraints),
		slog.String("package_version", v),
	)
	return p.Copy(), nil
}
