package expr_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/expr"
)

func TestCompileVersionFilter(t *testing.T) {
	t.Parallel()
	data := []struct {
		title         string
		versionFilter string
		isErr         bool
	}{
		{
			title:         "normal",
			versionFilter: `semver(">= 1.0.0")`,
		},
		{
			title:         "invalid version constraint",
			versionFilter: `>= v1.0.0`,
			isErr:         true,
		},
	}

	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			prog, err := expr.CompileVersionFilter(d.versionFilter)
			if d.isErr {
				if err == nil {
					t.Fatal("err should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if prog == nil {
				t.Fatal("prog must not be nil")
			}
		})
	}
}
