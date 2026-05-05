package expr

func emptySemver(_ string) bool {
	return false
}

func EvaluateVersionConstraints(constraint, v, semver string) (bool, error) {
	return evaluateBool(constraint, map[string]any{
		keyVersion:           "",
		"SemVer":             "",
		keySemver:            emptySemver,
		keySemverWithVersion: compare,
	}, map[string]any{
		keyVersion:           v,
		"SemVer":             semver,
		keySemver:            getCompareFunc(semver),
		keySemverWithVersion: compare,
	})
}
