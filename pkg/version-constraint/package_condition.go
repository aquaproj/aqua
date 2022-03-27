package constraint

import (
	"fmt"
	"runtime"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
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

func (pkgCondition *PackageCondition) Raw() string {
	return pkgCondition.raw
}

func (pkgCondition *PackageCondition) Compile() error {
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
	if err := pkgCondition.Compile(); err != nil {
		return false, err
	}
	a, err := expr.Run(pkgCondition.expr, map[string]interface{}{
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
	})
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	return a.(bool), nil
}

func (pkgCondition *PackageCondition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	pkgCondition.raw = raw
	return nil
}
