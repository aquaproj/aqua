package constraint

import (
	"fmt"

	"github.com/antonmedv/expr"
)

func evaluateBool(expression string, env, input interface{}) (bool, error) {
	compiled, err := expr.Compile(expression, expr.AsBool(), expr.Env(env))
	if err != nil {
		return false, fmt.Errorf("parse the expression: %w", err)
	}
	a, err := expr.Run(compiled, input)
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	f, ok := a.(bool)
	if !ok {
		return false, errMustBeBoolean
	}
	return f, nil
}
