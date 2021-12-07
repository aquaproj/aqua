package controller

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/hashicorp/go-version"
)

type VersionConstraints struct {
	raw  string
	expr *vm.Program
}

func NewVersionConstraints(s string) *VersionConstraints {
	return &VersionConstraints{
		raw: s,
	}
}

func (constraints *VersionConstraints) Compile() error {
	if constraints.expr != nil {
		return nil
	}
	a, err := expr.Compile(constraints.raw, expr.AsBool(), expr.Env(map[string]interface{}{
		"semver": func(s string) bool {
			return false
		},
	}))
	if err != nil {
		return fmt.Errorf("parse constraints: %w", err)
	}
	constraints.expr = a
	return nil
}

func (constraints *VersionConstraints) Check(v string) (bool, error) {
	if err := constraints.Compile(); err != nil {
		return false, err
	}
	a, err := expr.Run(constraints.expr, map[string]interface{}{
		"Version": v,
		"semver": func(s string) bool {
			a, err := version.NewConstraint(s)
			if err != nil {
				panic(err)
			}
			ver, err := version.NewVersion(v)
			if err != nil {
				panic(err)
			}
			return a.Check(ver)
		},
	})
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	return a.(bool), nil
}

func (constraints *VersionConstraints) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	constraints.raw = raw
	return nil
}
