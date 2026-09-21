package versiongetter_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
)

func TestFuzzyGetter_Get(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		pkg            *registry.PackageInfo
		currentVersion string
		useFinder      bool
		version        string
		idxs           []int
		versions       map[string][]*fuzzyfinder.Item
	}{
		{
			name: "normal",
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			currentVersion: "v2.0.0",
			version:        "v4.6.0",
			versions: map[string][]*fuzzyfinder.Item{
				"suzuki-shunsuke/tfcmt": {
					{
						Item: "v4.6.0",
					},
					{
						Item: "v3.0.0",
					},
					{
						Item: "v2.0.0",
					},
				},
			},
		},
		{
			name: "finder",
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			useFinder:      true,
			idxs:           []int{1},
			currentVersion: "v2.0.0",
			version:        "v3.0.0",
			versions: map[string][]*fuzzyfinder.Item{
				"suzuki-shunsuke/tfcmt": {
					{
						Item: "v4.6.0",
					},
					{
						Item: "v3.0.0",
					},
					{
						Item: "v2.0.0",
					},
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			finder := fuzzyfinder.NewMock(d.idxs, nil)
			vg := versiongetter.NewMockVersionGetter(d.versions)
			// A registry other than the standard one keeps asking upstream, which
			// is the getter under test here.
			fg := versiongetter.NewFuzzy(finder, vg, nil)
			version := fg.Get(t.Context(), logger, "foo", d.pkg, d.currentVersion, d.useFinder, -1)
			if version != d.version {
				t.Fatalf("wanted %s, got %s", d.version, version)
			}
		})
	}
}

// Where the versions come from is decided by the registry. A package from the
// standard registry is offered what aqua-registry-g2 can lock; one from any other
// registry keeps asking upstream, because it has no branch there.
func TestFuzzyGetter_Get_source(t *testing.T) {
	t.Parallel()
	pkg := &registry.PackageInfo{Name: "cli/cli"}
	upstream := versiongetter.NewMockVersionGetter(map[string][]*fuzzyfinder.Item{
		"cli/cli": fuzzyfinder.ConvertStringsToItems([]string{"v3.0.0"}),
	})
	fg := versiongetter.NewFuzzy(fuzzyfinder.NewMock([]int{0}, nil), upstream, newG2("v2.0.0"))
	logger := slog.New(slog.DiscardHandler)

	if got := fg.Get(t.Context(), logger, "standard", pkg, "", false, -1); got != "v2.0.0" {
		t.Errorf("the standard registry got %q, want v2.0.0 from g2", got)
	}
	if got := fg.Get(t.Context(), logger, "foo", pkg, "", false, -1); got != "v3.0.0" {
		t.Errorf("another registry got %q, want v3.0.0 from upstream", got)
	}
}
