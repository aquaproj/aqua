package expr

func emptySemver(s string) bool {
	return false
}

func EvaluateVersionConstraints(constraint, v, semver string) (bool, error) {
	return evaluateBool(constraint, map[string]interface{}{
		"Version":           "",
		"SemVer":            "",
		"semver":            emptySemver,
		"semverWithVersion": compare,
	}, map[string]interface{}{
		"Version":           v,
		"SemVer":            semver,
		"semver":            getCompareFunc(semver),
		"semverWithVersion": compare,
	})
}
