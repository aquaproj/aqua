package expr

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func CompileAssetFilter(assetFilter string) (*vm.Program, error) {
	return expr.Compile(assetFilter, expr.AsBool(), expr.Env(map[string]any{ //nolint:wrapcheck
		"Asset": "",
	}))
}

func EvaluateAssetFilter(prog *vm.Program, asset string) (bool, error) {
	return evaluateBoolProg(prog, map[string]any{
		"Asset": asset,
	})
}
