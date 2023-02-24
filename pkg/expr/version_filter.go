package expr

import (
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

func CompileVersionFilter(versionFilter string) (*vm.Program, error) {
	return expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]interface{}{ //nolint:wrapcheck
		"Version":           "",
		"semver":            emptySemver,
		"semverWithVersion": emptySemverWithVersion,
		"trimPrefix":        emptyTrimPrefix,
	}))
}

func CompileVersionFilterForTest(versionFilter string) *vm.Program {
	p, err := expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]interface{}{
		"Version":           "",
		"semver":            emptySemver,
		"semverWithVersion": emptySemverWithVersion,
		"trimPrefix":        emptyTrimPrefix,
	}))
	if err != nil {
		panic(err)
	}
	return p
}

func EvaluateVersionFilter(prog *vm.Program, v string) (bool, error) {
	return evaluateBoolProg(prog, map[string]interface{}{
		"Version":           v,
		"semver":            getSemverFunc(v),
		"semverWithVersion": semverWithVersion,
		"trimPrefix":        strings.TrimPrefix,
	})
}
