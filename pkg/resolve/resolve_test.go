package resolve_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
	"github.com/google/go-cmp/cmp"
)

func logger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// envs describes each entry by the environment it is for, which is what the entries
// are meant to differ by.
func envs(pkgs []*lockfile.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.OS + "/" + p.Arch
		if libc := p.Variants["libc"]; libc != "" {
			out[i] += "/" + libc
		}
	}
	return out
}

func TestResolve(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "cli/cli",
		Version: "v2.1.0",
		PkgInfo: &registry.PackageInfo{
			Name:          "cli/cli",
			Type:          "github_release",
			RepoOwner:     "cli",
			RepoName:      "cli",
			Asset:         "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
			Format:        "tar.gz",
			Files:         []*registry.File{{Name: "gh", Src: "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}/bin/gh"}},
			SupportedEnvs: registry.SupportedEnvs{"darwin", "linux"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}

	want := &lockfile.Package{
		Name:      "cli/cli",
		Version:   "v2.1.0",
		OS:        "linux",
		Arch:      "amd64",
		Type:      "github_release",
		RepoOwner: "cli",
		RepoName:  "cli",
		Asset:     "gh_2.1.0_linux_amd64.tar.gz",
		Format:    "tar.gz",
		Files:     []*lockfile.File{{Name: "gh", Src: "gh_2.1.0_linux_amd64/bin/gh"}},
	}
	if diff := cmp.Diff(want, pkgs[2]); diff != "" {
		t.Errorf("the entry is wrong (-want +got):\n%s", diff)
	}
}

// An override that differs only by libc produces a sibling entry rather than
// replacing the one it shares an os and arch with.
func TestResolve_variants(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
			Overrides: []*registry.Override{
				{
					GOOS:     "linux",
					Variants: []*registry.Variant{{Key: "libc", Value: "musl"}},
					Asset:    "foo_{{.OS}}_{{.Arch}}_musl.tar.gz",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"linux/amd64", "linux/amd64/musl"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("foo_linux_amd64.tar.gz", pkgs[0].Asset); diff != "" {
		t.Errorf("the default asset is wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("foo_linux_amd64_musl.tar.gz", pkgs[1].Asset); diff != "" {
		t.Errorf("the musl asset is wrong (-want +got):\n%s", diff)
	}
}

// Windows overrides change where the executable sits, so files have to be resolved
// per environment rather than once from the top level.
func TestResolve_windowsFiles(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.{{.Format}}",
			Format:        "tar.gz",
			Files:         []*registry.File{{Name: "foo", Src: "foo_{{.OS}}_{{.Arch}}/foo"}},
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64", "windows/amd64"},
			Overrides: []*registry.Override{
				{GOOS: "windows", Format: "zip", Files: []*registry.File{{Name: "foo", Src: "bin/foo"}}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, p := range pkgs {
		got[p.OS] = p.Files[0].Src
	}
	// Windows takes the override's path and gains the extension aqua adds there.
	if diff := cmp.Diff(map[string]string{"linux": "foo_linux_amd64/foo", "windows": "bin/foo.exe"}, got); diff != "" {
		t.Errorf("the files are wrong (-want +got):\n%s", diff)
	}
}

func TestResolve_http(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "nodejs/node",
		Version: "v20.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "http",
			URL:           "https://nodejs.org/dist/{{.Version}}/node-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("https://nodejs.org/dist/v20.0.0/node-v20.0.0-linux-amd64.tar.gz", pkgs[0].URL); diff != "" {
		t.Errorf("the URL is wrong (-want +got):\n%s", diff)
	}
}

// A package that resolves to nothing anywhere can't be locked, and an empty list
// would look like a successful resolution.
func TestResolve_noSupportedEnv(t *testing.T) {
	t.Parallel()
	_, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		},
	})
	if err == nil {
		t.Fatal("an error must be returned")
	}
}

// An override that declares a variant without changing anything produces an entry
// identical to the unconstrained one. A machine finding no entry for its libc falls
// through to the unconstrained entry and installs the same thing, so keeping both
// writes the same answer twice.
func TestResolve_dropRedundantVariants(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
			Overrides: []*registry.Override{
				// glibc changes nothing, so its entry is the unconstrained one.
				{GOOS: "linux", Variants: []*registry.Variant{{Key: "libc", Value: "glibc"}}},
				{
					GOOS:     "linux",
					Variants: []*registry.Variant{{Key: "libc", Value: "musl"}},
					Asset:    "foo_{{.OS}}_{{.Arch}}_musl.tar.gz",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"linux/amd64", "linux/amd64/musl"}, envs(pkgs)); diff != "" {
		t.Errorf("the environments are wrong (-want +got):\n%s", diff)
	}
}

// The entry kept is the unconstrained one, even when the package's base is the musl
// build and the override is the glibc one. Dropping it would leave nothing for a musl
// machine, which is the one the base was written for.
func TestResolve_keepDifferingVariants(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}_musl.tar.gz",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
			Overrides: []*registry.Override{
				{
					GOOS:     "linux",
					Variants: []*registry.Variant{{Key: "libc", Value: "glibc"}},
					Asset:    "foo_{{.OS}}_{{.Arch}}_gnu.tar.gz",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"linux/amd64", "linux/amd64/glibc"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff("foo_linux_amd64_musl.tar.gz", pkgs[0].Asset); diff != "" {
		t.Errorf("the fallback asset is wrong (-want +got):\n%s", diff)
	}
}

// TestResolve_varDefault covers a definition that writes part of its URL as a
// variable. Nothing supplies one here, so the default in the definition is all there
// is; without it the template renders "<no value>" and the entry points at a URL that
// answers 404.
func TestResolve_varDefault(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "flutter/flutter",
		Version: "3.47.5",
		PkgInfo: &registry.PackageInfo{
			Name:      "flutter/flutter",
			Type:      "http",
			RepoOwner: "flutter",
			RepoName:  "flutter",
			URL:       "https://example.com/releases/{{.Vars.channel}}/{{.OS}}/flutter_{{.OS}}_{{trimV .Version}}-{{.Vars.channel}}.{{.Format}}",
			Format:    "zip",
			Vars: []*registry.Var{
				{Name: "channel", Default: "stable"},
			},
			Replacements:  map[string]string{"darwin": "macos"},
			SupportedEnvs: []string{"darwin/arm64"},
		},
		SupportedEnvs: []string{"darwin/arm64"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("got %d entries, want 1", len(pkgs))
	}
	want := "https://example.com/releases/stable/macos/flutter_macos_3.47.5-stable.zip"
	if diff := cmp.Diff(want, pkgs[0].URL); diff != "" {
		t.Error(diff)
	}
}
