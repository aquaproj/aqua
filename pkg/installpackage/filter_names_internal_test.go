package installpackage

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func Test_matchNames(t *testing.T) {
	t.Parallel()
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "cli/cli",
			Version: "v2.0.0",
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "cli",
			RepoName:  "cli",
			Files: []*registry.File{
				{Name: "gh"},
			},
		},
	}
	data := []struct {
		name  string
		names map[string]struct{}
		exp   []string
	}{
		{
			name:  "package name",
			names: map[string]struct{}{"cli/cli": {}},
			exp:   []string{"cli/cli"},
		},
		{
			name:  "command name",
			names: map[string]struct{}{"gh": {}},
			exp:   []string{"gh"},
		},
		{
			name:  "package name and command name",
			names: map[string]struct{}{"gh": {}, "cli/cli": {}, "foo": {}},
			exp:   []string{"cli/cli", "gh"},
		},
		{
			name:  "not match",
			names: map[string]struct{}{"foo": {}},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(d.exp, matchNames(pkg, d.names)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
