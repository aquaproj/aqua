package expr_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/expr"
)

func TestVersionConstraints_Check(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title       string
		constraints string
		version     string
		semver      string
		exp         bool
		isErr       bool
	}{
		{
			title:       "true",
			constraints: `semver(">= 0.4.0")`,
			version:     "v0.4.0",
			semver:      "v0.4.0",
			exp:         true,
		},
		{
			title:       "false",
			constraints: `semver(">= 0.4.0")`,
			version:     "v0.3.0",
			semver:      "v0.3.0",
			exp:         false,
		},
		{
			title:       "semverWithVersion true",
			constraints: `semverWithVersion(">= 4.2.0", trimPrefix(Version, "kustomize/"))`,
			version:     "kustomize/v4.3.0",
			semver:      "v4.3.0",
			exp:         true,
		},
		{
			title:       "semverWithVersion false",
			constraints: `semverWithVersion(">= 4.2.0", trimPrefix(Version, "kustomize/"))`,
			version:     "kustomize/v0.3.0",
			semver:      "v0.3.0",
			exp:         false,
		},
		{
			title:       "invalid expression",
			constraints: `>= 0.4.0`,
			version:     "v0.3.0",
			semver:      "v0.3.0",
			isErr:       true,
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			b, err := expr.EvaluateVersionConstraints(d.constraints, d.version, d.semver)
			if d.isErr {
				if err == nil {
					t.Fatal("err should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if b != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, b)
			}
		})
	}
}
