package resolve_test

import (
	"log/slog"
	"strings"
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

// A release on a forge instance is a repository per instance, so the entry says which
// one: without it, installing would read the same owner and name on github.com.
func TestResolve_forgeInstance(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "codeberg.org/mergiraf/mergiraf",
		Version: "v0.20.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "forgejo_release",
			Host:          "codeberg.org",
			RepoOwner:     "mergiraf",
			RepoName:      "mergiraf",
			Asset:         "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pkgs[0].Host != "codeberg.org" {
		t.Errorf("the host is %q, want the instance the release is on", pkgs[0].Host)
	}
}

// A gitlab_release package is on gitlab.com unless it says otherwise, and the entry says
// which instance even then: what reads it reads nothing else.
func TestResolve_gitLabDefaultHost(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "gitlab.com/gitlab-org/cli",
		Version: "v1.122.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "gitlab_release",
			RepoOwner:     "gitlab-org",
			RepoName:      "cli",
			Asset:         "glab_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
			Format:        "tar.gz",
			SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pkgs[0].Host != "gitlab.com" {
		t.Errorf("the host is %q, want the instance a gitlab_release package is on", pkgs[0].Host)
	}
}

// A github_archive package is the repository's archive of the tag, which needs no asset
// or URL to say where it is: bats-core/bats-core.
func TestResolve_githubArchive(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "bats-core/bats-core",
		Version: "v1.14.0",
		PkgInfo: &registry.PackageInfo{
			Type: "github_archive", RepoOwner: "bats-core", RepoName: "bats-core",
			SupportedEnvs: registry.SupportedEnvs{"darwin", "linux"},
			Files:         []*registry.File{{Name: "bats", Src: "bats-core-{{trimV .Version}}/bin/bats"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 4 {
		t.Fatalf("got %d entries, want one for each of darwin and linux on amd64 and arm64", len(pkgs))
	}
	for _, p := range pkgs {
		if p.Type != "github_archive" || len(p.Files) != 1 || p.Files[0].Src != "bats-core-1.14.0/bin/bats" {
			t.Errorf("%s/%s is wrong: %+v", p.OS, p.Arch, p)
		}
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

// TestResolve_varWithoutValue covers a variable nothing can fill. ApplyVars refuses
// only a required one, so this would otherwise be rendered as "<no value>" into the
// URL and served as the answer.
func TestResolve_varWithoutValue(t *testing.T) {
	t.Parallel()
	_, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "example/example",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Name:      "example/example",
			Type:      "http",
			RepoOwner: "example",
			RepoName:  "example",
			URL:       "https://example.com/{{.Vars.channel}}/example.zip",
			Format:    "zip",
			Vars: []*registry.Var{
				{Name: "channel"},
			},
			SupportedEnvs: []string{"darwin/arm64"},
		},
		SupportedEnvs: []string{"darwin/arm64"},
	})
	if err == nil {
		t.Fatal("want an error, got none")
	}
	if !strings.Contains(err.Error(), "channel") {
		t.Errorf("the error doesn't name the variable: %v", err)
	}
}

// A release with only an amd64 Windows build runs on a Windows arm64 machine under
// emulation, and aqua installs it there although supported_envs names windows/amd64
// alone: jacobdeichert/mask 0.11.5.
func TestResolve_windowsARMEmulation(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "jacobdeichert/mask",
		Version: "mask/0.11.5",
		PkgInfo: &registry.PackageInfo{
			Type:                "github_release",
			RepoOwner:           "jacobdeichert",
			RepoName:            "mask",
			Asset:               "mask-0.11.5-{{.Arch}}-{{.OS}}.{{.Format}}",
			Format:              "zip",
			WindowsARMEmulation: true,
			Replacements:        map[string]string{"amd64": "x86_64", "arm64": "aarch64", "darwin": "apple-darwin", "linux": "unknown-linux-musl", "windows": "pc-windows-msvc"},
			SupportedEnvs:       registry.SupportedEnvs{"linux/amd64", "darwin/arm64", "windows/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"linux/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
	if got := pkgs[3].Asset; got != "mask-0.11.5-x86_64-pc-windows-msvc.zip" {
		t.Errorf("windows/arm64 resolves to %s, want the amd64 build", got)
	}
	// The entry says it runs under emulation, and the native amd64 one doesn't.
	if !pkgs[3].WindowsARMEmulation {
		t.Error("windows/arm64 doesn't say it runs under emulation")
	}
	if pkgs[2].WindowsARMEmulation || pkgs[3].Rosetta2 {
		t.Errorf("a flag is set where emulation doesn't apply: %+v %+v", pkgs[2], pkgs[3])
	}
}

// supported_envs in the configuration still limits the emulated environment.
func TestResolve_emulationLimitedByConfig(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName:       "foo/foo",
		Version:       "v1.0.0",
		SupportedEnvs: []string{"windows/amd64"},
		PkgInfo: &registry.PackageInfo{
			Type:                "github_release",
			RepoOwner:           "foo",
			RepoName:            "foo",
			Asset:               "foo_{{.OS}}_{{.Arch}}.zip",
			WindowsARMEmulation: true,
			SupportedEnvs:       registry.SupportedEnvs{"windows/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"windows/amd64"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
}

// A configuration that allows the arm64 environment gets it.
func TestResolve_emulationAllowedByConfig(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName:       "foo/foo",
		Version:       "v1.0.0",
		SupportedEnvs: []string{"windows"},
		PkgInfo: &registry.PackageInfo{
			Type:                "github_release",
			RepoOwner:           "foo",
			RepoName:            "foo",
			Asset:               "foo_{{.OS}}_{{.Arch}}.zip",
			WindowsARMEmulation: true,
			SupportedEnvs:       registry.SupportedEnvs{"windows/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"windows/amd64", "windows/arm64"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
}

// Rosetta 2 is the same on macOS: darwin/arm64 runs the amd64 build.
func TestResolve_rosetta2(t *testing.T) {
	t.Parallel()
	pkgs, err := resolve.Resolve(logger(), &resolve.Param{
		PkgName: "foo/foo",
		Version: "v1.0.0",
		PkgInfo: &registry.PackageInfo{
			Type:          "github_release",
			RepoOwner:     "foo",
			RepoName:      "foo",
			Asset:         "foo_{{.OS}}_{{.Arch}}.tar.gz",
			Rosetta2:      true,
			SupportedEnvs: registry.SupportedEnvs{"darwin/amd64"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"darwin/amd64", "darwin/arm64"}, envs(pkgs)); diff != "" {
		t.Fatalf("the environments are wrong (-want +got):\n%s", diff)
	}
	if got := pkgs[1].Asset; got != "foo_darwin_amd64.tar.gz" {
		t.Errorf("darwin/arm64 resolves to %s, want the amd64 build", got)
	}
	if !pkgs[1].Rosetta2 || pkgs[0].Rosetta2 {
		t.Errorf("only darwin/arm64 runs under Rosetta 2: amd64=%v arm64=%v", pkgs[0].Rosetta2, pkgs[1].Rosetta2)
	}
}

// A definition saying rosetta2 doesn't make an entry depend on it: a universal build or
// a package built from source runs on arm64 as it is.
func TestResolve_rosetta2NotUsed(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		info *registry.PackageInfo
	}{
		{
			name: "universal build",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "foo", RepoName: "foo",
				Asset:         "foo_darwin_universal.tar.gz",
				Rosetta2:      true,
				SupportedEnvs: registry.SupportedEnvs{"darwin"},
			},
		},
		{
			name: "go_install",
			info: &registry.PackageInfo{
				Type: "go_install", Path: "example.com/foo",
				Rosetta2:      true,
				SupportedEnvs: registry.SupportedEnvs{"darwin"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkgs, err := resolve.Resolve(logger(), &resolve.Param{PkgName: "foo/foo", Version: "v1.0.0", PkgInfo: tt.info})
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range pkgs {
				if p.Rosetta2 {
					t.Errorf("%s/%s says it runs under Rosetta 2", p.OS, p.Arch)
				}
			}
		})
	}
}

// An amd64 build whose name doesn't change with the architecture still runs under
// Rosetta 2 on an arm64 Mac: Songmu/gocredits names the architecture in the template.
func TestResolve_rosetta2NameSaysAmd64(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		info *registry.PackageInfo
	}{
		{
			name: "the template names amd64",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "Songmu", RepoName: "gocredits",
				Asset:         "gocredits_{{.Version}}_{{.OS}}_amd64.{{.Format}}",
				Format:        "zip",
				Rosetta2:      true,
				SupportedEnvs: registry.SupportedEnvs{"darwin"},
			},
		},
		{
			name: "only amd64 is supported",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "foo", RepoName: "foo",
				Asset:         "foo_darwin.tar.gz",
				Rosetta2:      true,
				SupportedEnvs: registry.SupportedEnvs{"darwin/amd64"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkgs, err := resolve.Resolve(logger(), &resolve.Param{PkgName: "foo/foo", Version: "v1.0.0", PkgInfo: tt.info})
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range pkgs {
				if want := p.Arch == "arm64"; p.Rosetta2 != want {
					t.Errorf("%s/%s: rosetta2 is %v, want %v", p.OS, p.Arch, p.Rosetta2, want)
				}
			}
		})
	}
}

// Each OS's flag is read from its own declaration, and the name decides only by whole
// words.
func TestResolve_emulationPerOS(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                     string
		info                     *registry.PackageInfo
		rosetta2, windowsARMEmul bool
	}{
		{
			name: "windows_arm_emulation says nothing about macOS",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "o", RepoName: "r",
				Asset:               "r_{{.OS}}_amd64.zip",
				WindowsARMEmulation: true,
				SupportedEnvs:       registry.SupportedEnvs{"darwin", "windows"},
			},
			windowsARMEmul: true,
		},
		{
			name: "a tool named x64 isn't an amd64 build",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "o", RepoName: "x64tool",
				Asset:         "x64tool_{{.OS}}.zip",
				Rosetta2:      true,
				SupportedEnvs: registry.SupportedEnvs{"darwin"},
			},
		},
		{
			name: "a darwin override naming the amd64 build",
			info: &registry.PackageInfo{
				Type: "github_release", RepoOwner: "o", RepoName: "r",
				Asset:    "r_{{.OS}}_{{.Arch}}.tar.gz",
				Rosetta2: true,
				Overrides: []*registry.Override{
					{GOOS: "darwin", Asset: "r_macos_x86_64.tar.gz"},
				},
				SupportedEnvs: registry.SupportedEnvs{"darwin"},
			},
			rosetta2: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkgs, err := resolve.Resolve(logger(), &resolve.Param{PkgName: "o/r", Version: "v1.0.0", PkgInfo: tt.info})
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range pkgs {
				if p.Arch != "arm64" {
					continue
				}
				want := tt.rosetta2
				got := p.Rosetta2
				if p.OS == "windows" {
					want, got = tt.windowsARMEmul, p.WindowsARMEmulation
				}
				if got != want {
					t.Errorf("%s/%s: flag is %v, want %v", p.OS, p.Arch, got, want)
				}
			}
		})
	}
}
