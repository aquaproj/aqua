package expr

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func evaluateBool(expression string, env, input any) (bool, error) {
	compiled, err := expr.Compile(expression, expr.AsBool(), expr.Env(env))
	if err != nil {
		return false, fmt.Errorf("parse the expression: %w", err)
	}
	return evaluateBoolProg(compiled, input)
}

func evaluateBoolProg(prog *vm.Program, input any) (bool, error) {
	a, err := expr.Run(prog, input)
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	f, ok := a.(bool)
	if !ok {
		return false, errMustBeBoolean
	}
	return f, nil
}
