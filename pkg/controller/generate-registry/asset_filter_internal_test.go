package genrgst

import (
	"log/slog"
	"testing"
)

func discard() *slog.Logger { return slog.New(slog.DiscardHandler) }

// Which of a release's assets is the package can change over its history.
// openai/codex is the one this was written for: codex was codex-<triple> until 0.133
// and codex-package-<triple> after it, in releases that carry both names and about a
// dozen other programs besides.
func TestConfig_assetFilter(t *testing.T) {
	t.Parallel()
	cfg := &Config{VersionPrefix: "rust-"}
	if err := cfg.FromRaw(&RawConfig{
		Package:         "openai/codex",
		VersionPrefix:   "rust-",
		AllAssetsFilter: `Asset matches "^never-"`,
		VersionOverrides: []*RawVersionOverride{
			{
				VersionConstraint: `semver("> 0.133.0-alpha.1")`,
				AllAssetsFilter:   `Asset matches "^codex-package-"`,
			},
			{
				VersionConstraint: "true",
				AllAssetsFilter:   `Asset matches "^codex-(x86_64|aarch64)-"`,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	for _, d := range []struct {
		tag   string
		asset string
		want  bool // excluded
	}{
		// The newer releases keep only the package archive.
		{tag: "rust-v0.156.1", asset: "codex-package-x86_64-apple-darwin.tar.gz", want: false},
		{tag: "rust-v0.156.1", asset: "codex-x86_64-apple-darwin.tar.gz", want: true},
		{tag: "rust-v0.156.1", asset: "bwrap-x86_64-unknown-linux-musl.tar.gz", want: true},
		// The older ones have no package archive at all, and keep the flat binary.
		{tag: "rust-v0.19.0", asset: "codex-x86_64-apple-darwin.tar.gz", want: false},
		{tag: "rust-v0.19.0", asset: "codex-exec-x86_64-apple-darwin.tar.gz", want: true},
	} {
		t.Run(d.tag+"/"+d.asset, func(t *testing.T) {
			t.Parallel()
			if got := excludeAsset(discard(), d.tag, d.asset, cfg); got != d.want {
				t.Errorf("excluded is %v, want %v", got, d.want)
			}
		})
	}
}

// A configuration naming no version_overrides answers with the top level, which is
// what every package that has one does today.
func TestConfig_assetFilter_topLevelOnly(t *testing.T) {
	t.Parallel()
	cfg := &Config{}
	if err := cfg.FromRaw(&RawConfig{
		Package:         "cubefs/cubefs",
		AllAssetsFilter: `Asset matches "^cubefs-"`,
	}); err != nil {
		t.Fatal(err)
	}
	if excludeAsset(discard(), "v3.5.0", "cubefs-server.tar.gz", cfg) {
		t.Error("the package's own asset was excluded")
	}
	if !excludeAsset(discard(), "v3.5.0", "something-else.tar.gz", cfg) {
		t.Error("an asset the filter rejects was kept")
	}
}
