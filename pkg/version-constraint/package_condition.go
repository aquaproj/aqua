package constraint

import (
	"fmt"
	"runtime"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/invopop/jsonschema"
)

type PackageCondition struct {
	raw  string
	expr *vm.Program
}

func NewPackageCondition(s string) *PackageCondition {
	return &PackageCondition{
		raw: s,
	}
}

func (PackageCondition) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Description: "expr's expression. The evaluation result must be a boolean. If the evaluation result is false, aqua skips installing the package and outputs the debug log. If supported_if isn't set, the package is always installed.",
	}
}

func (pkgCondition *PackageCondition) Raw() string {
	return pkgCondition.raw
}

func (pkgCondition *PackageCondition) compile() error {
	if pkgCondition.expr != nil {
		return nil
	}
	a, err := expr.Compile(pkgCondition.raw, expr.AsBool(), expr.Env(map[string]interface{}{
		"GOOS":   "",
		"GOARCH": "",
	}))
	if err != nil {
		return fmt.Errorf("parse constraints: %w", err)
	}
	pkgCondition.expr = a
	return nil
}

func (pkgCondition *PackageCondition) Check() (bool, error) {
	if err := pkgCondition.compile(); err != nil {
		return false, err
	}
	a, err := expr.Run(pkgCondition.expr, map[string]interface{}{
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
	})
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	f, ok := a.(bool)
	if !ok {
		return false, errPackageConditionMustBeBoolean
	}
	return f, nil
}

func (pkgCondition *PackageCondition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	pkgCondition.raw = raw
	return nil
}
