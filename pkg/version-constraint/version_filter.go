package constraint

import (
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

func CompileVersionFilter(versionFilter string) (*vm.Program, error) {
	return expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]interface{}{ //nolint:wrapcheck
		"Version": "",
		"semver": func(s string) bool {
			return false
		},
		"semverWithVersion": func(constr, ver string) bool {
			return false
		},
		"trimPrefix": func(s, prefix string) string {
			return ""
		},
	}))
}

func EvaluateVersionFilter(prog *vm.Program, v string) (bool, error) {
	return evaluateBoolProg(prog, map[string]interface{}{
		"Version":           v,
		"semver":            getSemverFunc(v),
		"semverWithVersion": semverWithVersion,
		"trimPrefix":        strings.TrimPrefix,
	})
}
