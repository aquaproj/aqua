package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

// A release may write a platform in a way the parser doesn't know, and then the
// asset belongs to no platform at all: luau-lang/luau calls its Linux build
// luau-ubuntu.zip, so the package was generated with no Linux in it.
func TestParseAssetName_spellings(t *testing.T) {
	t.Parallel()
	got := asset.ParseAssetName("luau-ubuntu.zip", "0.739")
	if got.OS != "" {
		t.Errorf("a spelling the parser doesn't know was read as %q", got.OS)
	}

	got = asset.ParseAssetName("luau-ubuntu.zip", "0.739", asset.Spellings{"linux": "ubuntu"})
	if got.OS != "linux" {
		t.Errorf("the OS is %q, want linux", got.OS)
	}
	// An architecture the name doesn't mention is amd64, as it is for any other
	// asset.
	if got.Arch != "amd64" {
		t.Errorf("the arch is %q, want amd64", got.Arch)
	}
	if got.Template != "luau-{{.OS}}.{{.Format}}" {
		t.Errorf("the template is %q", got.Template)
	}
	// The template renders back to the name the release actually uses.
	if got.Replacements["linux"] != "ubuntu" {
		t.Errorf("the replacement is %q, want ubuntu", got.Replacements["linux"])
	}
}

// A spelling given by name is more specific than one the parser guessed at, so it is
// tried first.
func TestParseAssetName_spellingsWin(t *testing.T) {
	t.Parallel()
	got := asset.ParseAssetName("tool-mac-x64.tar.gz", "v1.0.0", asset.Spellings{"darwin": "mac", "amd64": "x64"})
	if got.OS != "darwin" || got.Arch != "amd64" {
		t.Errorf("got %s/%s, want darwin/amd64", got.OS, got.Arch)
	}
}

// Nothing is said about a platform the definition doesn't mention.
func TestParseAssetName_spellingsEmpty(t *testing.T) {
	t.Parallel()
	got := asset.ParseAssetName("tool-linux-amd64.tar.gz", "v1.0.0", asset.Spellings{})
	if got.OS != "linux" || got.Arch != "amd64" {
		t.Errorf("got %s/%s, want linux/amd64", got.OS, got.Arch)
	}
}
