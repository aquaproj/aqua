package generate

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
)

func Test_filterRelease(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		release *github.RepositoryRelease
		filters []*Filter
		exp     bool
	}{
		{
			name: "no filter",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("v1.0.0"),
			},
			filters: []*Filter{
				{},
			},
			exp: true,
		},
		{
			name: "version_filter",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("v1.0.0"),
			},
			filters: []*Filter{
				{
					Filter: expr.CompileVersionFilterForTest(`Version startsWith "cli/"`),
				},
			},
			exp: false,
		},
		{
			name: "version_prefix",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("v1.0.0"),
			},
			filters: []*Filter{
				{
					Prefix: "cli/",
				},
			},
			exp: false,
		},
		{
			name: "version_constraints 1",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("v1.0.0"),
			},
			filters: []*Filter{
				{
					Prefix:     "",
					Constraint: `semver(">= 2.0.0")`,
				},
				{
					Prefix:     "cli/",
					Constraint: `true`,
				},
			},
			exp: false,
		},
		{
			name: "version_constraints 2",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("v2.0.0"),
			},
			filters: []*Filter{
				{
					Constraint: `semver(">= 2.0.0")`,
				},
				{
					Prefix:     "cli/",
					Constraint: `true`,
				},
			},
			exp: true,
		},
		{
			name: "version_constraints 3",
			release: &github.RepositoryRelease{
				TagName: ptr.StrP("cli/v1.0.0"),
			},
			filters: []*Filter{
				{
					Constraint: `semver(">= 2.0.0")`,
				},
				{
					Prefix:     "cli/",
					Constraint: `true`,
				},
			},
			exp: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := filterRelease(d.release, d.filters)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
