package expr

import (
	"log/slog"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func CompileVersionFilter(versionFilter string) (*vm.Program, error) {
	return expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]any{ //nolint:wrapcheck
		keyVersion:           "",
		keySemver:            emptySemver,
		keySemverWithVersion: emptySemverWithVersion,
	}))
}

func CompileVersionFilterForTest(versionFilter string) *vm.Program {
	p, err := expr.Compile(versionFilter, expr.AsBool(), expr.Env(map[string]any{
		keyVersion:           "",
		keySemver:            emptySemver,
		keySemverWithVersion: emptySemverWithVersion,
	}))
	if err != nil {
		panic(err)
	}
	return p
}

func EvaluateVersionFilter(logger *slog.Logger, prog *vm.Program, v string) (bool, error) {
	return evaluateBoolProg(prog, map[string]any{
		keyVersion:           v,
		keySemver:            getCompareFunc(logger, v),
		keySemverWithVersion: compareFunc(logger),
	})
}
