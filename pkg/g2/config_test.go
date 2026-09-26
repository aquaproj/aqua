package g2_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
)

// The overrides are ordered newest first, each carrying the lower bound of what it
// covers, so the first match is the right one.
func TestConfig_SetVersion(t *testing.T) {
	t.Parallel()
	cfg := &g2.Config{
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "tailwindlabs",
			RepoName:  "tailwindcss",
			VersionOverrides: []*registry.VersionOverride{
				{VersionConstraints: `semver("> 3.4.17")`, Asset: "newest"},
				{VersionConstraints: `semver("> 3.2.7")`, Asset: "3.4"},
				{VersionConstraints: `semver("> 3.0.7")`, Asset: "3.2"},
				{VersionConstraints: `semver("<= 3.0.7")`, Asset: "oldest"},
			},
		},
	}
	tests := map[string]string{
		"v4.0.0": "newest",
		"v3.5.0": "newest",
		"v3.3.0": "3.4",
		"v3.1.0": "3.2",
		"v3.0.1": "oldest",
		// Not a semver, so nothing matches and the newest definition answers.
		"nightly": "newest",
	}
	for v, want := range tests {
		t.Run(v, func(t *testing.T) {
			t.Parallel()
			got, err := cfg.SetVersion(slog.New(slog.DiscardHandler), v)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, got.Asset); diff != "" {
				t.Errorf("the definition is wrong (-want +got):\n%s", diff)
			}
			// The top level is the base the overrides inherit from.
			if got.RepoOwner != "tailwindlabs" {
				t.Errorf("the base wasn't inherited: %+v", got)
			}
		})
	}
}

// A file with no overrides says nothing about any version, so it is an error rather
// than the base standing in for one.
func TestConfig_SetVersion_noOverride(t *testing.T) {
	t.Parallel()
	cfg := &g2.Config{PackageInfo: &registry.PackageInfo{Type: "github_release"}}
	if _, err := cfg.SetVersion(slog.New(slog.DiscardHandler), "v1.0.0"); err == nil {
		t.Fatal("an error must be returned")
	}
}

// all_assets_filter has no place in aqua's own format, so it has to survive a round
// trip through this one.
func TestConfig_yaml(t *testing.T) {
	t.Parallel()
	b := []byte(`type: github_release
repo_owner: aubepkg
repo_name: aube
all_assets_filter: not (Asset matches "^libaube-")
version_overrides:
  - version_constraint: "true"
    asset: aube-{{.Version}}.tar.gz
`)
	cfg := &g2.Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`not (Asset matches "^libaube-")`, cfg.AllAssetsFilter); diff != "" {
		t.Errorf("the filter is wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("aubepkg", cfg.RepoOwner); diff != "" {
		t.Errorf("the base is wrong (-want +got):\n%s", diff)
	}
	if len(cfg.VersionOverrides) != 1 {
		t.Fatalf("got %d overrides, want 1", len(cfg.VersionOverrides))
	}
}
