package constraint

import (
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

type VersionFilter struct {
	raw  string
	expr *vm.Program
}

func NewVersionFilter(s string) *VersionFilter {
	return &VersionFilter{
		raw: s,
	}
}

func (vf *VersionFilter) Raw() string {
	return vf.raw
}

func (vf *VersionFilter) Compile() error {
	if vf.expr != nil {
		return nil
	}
	a, err := expr.Compile(vf.raw, expr.AsBool(), expr.Env(map[string]interface{}{
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
	if err != nil {
		return fmt.Errorf("parse constraints: %w", err)
	}
	vf.expr = a
	return nil
}

func (vf *VersionFilter) Check(v string) (bool, error) {
	if err := vf.Compile(); err != nil {
		return false, err
	}
	a, err := expr.Run(vf.expr, map[string]interface{}{
		"Version":           v,
		"semver":            getSemverFunc(v),
		"semverWithVersion": semverWithVersion,
		"trimPrefix":        strings.TrimPrefix,
	})
	if err != nil {
		return false, fmt.Errorf("evaluate the expression: %w", err)
	}
	return a.(bool), nil
}

func (vf *VersionFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	vf.raw = raw
	return nil
}
