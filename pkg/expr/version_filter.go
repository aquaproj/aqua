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
		"semverWithVersion": compare,
		"trimPrefix":        strings.TrimPrefix,
	}))
}

func CompileVersionFilterForTest(versionFilter string) *vm.Program {
	p, err := expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]interface{}{
		"Version":           "",
		"semver":            emptySemver,
		"semverWithVersion": compare,
		"trimPrefix":        strings.TrimPrefix,
	}))
	if err != nil {
		panic(err)
	}
	return p
}

func EvaluateVersionFilter(prog *vm.Program, v string) (bool, error) {
	return evaluateBoolProg(prog, map[string]interface{}{
		"Version":           v,
		"semver":            getCompareFunc(v),
		"semverWithVersion": compare,
		"trimPrefix":        strings.TrimPrefix,
	})
}
