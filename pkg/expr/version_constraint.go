package expr

import (
	"strings"
)

func emptySemver(s string) bool {
	return false
}

func EvaluateVersionConstraints(constraint, v, semver string) (bool, error) {
	return evaluateBool(constraint, map[string]interface{}{
		"Version":           "",
		"SemVer":            "",
		"semver":            emptySemver,
		"semverWithVersion": compare,
		"trimPrefix":        strings.TrimPrefix,
	}, map[string]interface{}{
		"Version": v,

		"SemVer": semver,

		"semver":            getCompareFunc(semver),
		"semverWithVersion": compare,

		"trimPrefix": strings.TrimPrefix,
	})
}
