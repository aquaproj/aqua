package expr

import "log/slog"

func emptySemver(_ string) bool {
	return false
}

func EvaluateVersionConstraints(logger *slog.Logger, constraint, v, semver string) (bool, error) {
	return evaluateBool(constraint, map[string]any{
		keyVersion:           "",
		"SemVer":             "",
		keySemver:            emptySemver,
		keySemverWithVersion: emptySemverWithVersion,
	}, map[string]any{
		keyVersion:           v,
		"SemVer":             semver,
		keySemver:            getCompareFunc(logger, semver),
		keySemverWithVersion: compareFunc(logger),
	})
}
