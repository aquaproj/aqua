package resolve_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
)

// The names of signature files are templates in a registry, the same way an asset
// name is, and are rendered for the environment the entry describes.
//
// An entry that kept them would need the replacements kept with it to expand them
// again, and would expand them wrongly without: ziglang/zig writes its signature as
// zig-{{.OS}}-{{.Arch}}-{{.Version}}.tar.xz.minisig and replaces darwin with macos
// and arm64 with aarch64.
func TestResolve_signingIsRendered(t *testing.T) {
	t.Parallel()
	url := "https://ziglang.org/download/{{.Version}}/zig-{{.Arch}}-{{.OS}}-{{.Version}}.tar.xz.minisig"
	pkgs, err := resolve.Resolve(slog.New(slog.DiscardHandler), &resolve.Param{
		PkgName: "ziglang/zig",
		Version: "0.15.2",
		PkgInfo: &registry.PackageInfo{
			Name:          "ziglang/zig",
			Type:          "http",
			URL:           "https://ziglang.org/download/{{.Version}}/zig-{{.Arch}}-{{.OS}}-{{.Version}}.tar.xz",
			Format:        "tar.xz",
			SupportedEnvs: registry.SupportedEnvs{"darwin/arm64"},
			Replacements:  registry.Replacements{"darwin": "macos", "arm64": "aarch64"},
			Minisign: &registry.Minisign{
				Type:      "http",
				URL:       &url,
				PublicKey: "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("got %d entries, want 1", len(pkgs))
	}
	want := "https://ziglang.org/download/0.15.2/zig-aarch64-macos-0.15.2.tar.xz.minisig"
	if got := pkgs[0].Minisign.URL; got == nil || *got != want {
		t.Errorf("the signature URL is %v, want %s", got, want)
	}
	// The entry must not be carrying the rule that produced it.
	if pkgs[0].Minisign.PublicKey == "" {
		t.Error("the public key was lost")
	}
}

// A package from a private repository says so, because that is what sends the
// download straight to the API instead of trying the anonymous URL first.
func TestResolve_private(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(slog.New(slog.DiscardHandler), &resolve.Param{
		PkgName: "example/private",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Name:          "example/private",
			Type:          "github_release",
			RepoOwner:     "example",
			RepoName:      "private",
			Asset:         "private_{{.OS}}_{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
			Private:       true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || !pkgs[0].Private {
		t.Fatalf("the entry doesn't say the repository is private: %+v", pkgs)
	}
	// The definition aqua installs from has to say it too, or the entry saying it
	// changes nothing.
	if !pkgs[0].PackageInfo().Private {
		t.Error("the definition built from the entry lost it")
	}
}
