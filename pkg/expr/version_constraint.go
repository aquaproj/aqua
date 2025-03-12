package expr

func emptySemver(_ string) bool {
	return false
}

func EvaluateVersionConstraints(constraint, v, semver string) (bool, error) {
	return evaluateBool(constraint, map[string]any{
		"Version":           "",
		"SemVer":            "",
		"semver":            emptySemver,
		"semverWithVersion": compare,
	}, map[string]any{
		"Version":           v,
		"SemVer":            semver,
		"semver":            getCompareFunc(semver),
		"semverWithVersion": compare,
	})
}
