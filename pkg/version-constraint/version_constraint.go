package constraint

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/hashicorp/go-version"
	"github.com/invopop/jsonschema"
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

func (VersionConstraints) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Description: "expr's expression. The evaluation result must be a boolean",
	}
}

func emptyTrimPrefix(string, string) string {
	return ""
}

func emptySemver(s string) bool {
	return false
}

func emptySemverWithVersion(constr, ver string) bool {
	return false
}

func (constraints *VersionConstraints) compile() error {
	if constraints.expr != nil {
		return nil
	}
	a, err := expr.Compile(constraints.raw, expr.AsBool(), expr.Env(map[string]interface{}{
		"Version":           "",
		"semver":            emptySemver,
		"semverWithVersion": emptySemverWithVersion,
		"trimPrefix":        emptyTrimPrefix,
	}))
	if err != nil {
		return fmt.Errorf("parse constraints: %w", err)
	}
	constraints.expr = a
	return nil
}

func getSemverFunc(v string) func(s string) bool {
	return func(s string) bool {
		a, err := version.NewConstraint(s)
		if err != nil {
			panic(err)
		}
		ver, err := version.NewVersion(v)
		if err != nil {
			panic(err)
		}
		return a.Check(ver)
	}
}

func semverWithVersion(constr, ver string) bool {
	a, err := version.NewConstraint(constr)
	if err != nil {
		panic(err)
	}
	v, err := version.NewVersion(ver)
	if err != nil {
		panic(err)
	}
	return a.Check(v)
}

func (constraints *VersionConstraints) Check(v string) (bool, error) {
	if err := constraints.compile(); err != nil {
		return false, err
	}
	return evaluateBoolProg(constraints.expr, map[string]interface{}{
		"Version":           v,
		"semver":            getSemverFunc(v),
		"semverWithVersion": semverWithVersion,
		"trimPrefix":        strings.TrimPrefix,
	})
}

func (constraints *VersionConstraints) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	constraints.raw = raw
	return nil
}

func (constraints *VersionConstraints) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("unmarshal version constraint as JSON: %w", err)
	}
	constraints.raw = raw
	return nil
}
