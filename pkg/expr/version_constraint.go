package expr

import (
	"strings"

	"github.com/hashicorp/go-version"
)

func emptyTrimPrefix(string, string) string {
	return ""
}

func emptySemver(s string) bool {
	return false
}

func emptySemverWithVersion(constr, ver string) bool {
	return false
}

func getSemverFunc(v string) func(s string) bool {
	return func(s string) bool {
		return semverWithVersion(s, v)
	}
}

func semverWithVersion(constr, ver string) bool {
	a, err := version.NewConstraint(constr)
	if err != nil {
		panic(err)
	}
	v, err := version.NewVersion(ver)
	if err != nil {
		panic(err)
	}
	return a.Check(v)
}

func EvaluateVersionConstraints(constraint, v string) (bool, error) {
	return evaluateBool(constraint, map[string]interface{}{
		"Version":           "",
		"semver":            emptySemver,
		"semverWithVersion": emptySemverWithVersion,
		"trimPrefix":        emptyTrimPrefix,
	}, map[string]interface{}{
		"Version":           v,
		"semver":            getSemverFunc(v),
		"semverWithVersion": semverWithVersion,
		"trimPrefix":        strings.TrimPrefix,
	})
}
