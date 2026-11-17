package lockresolve_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
)

// fakeG2 answers with the definitions it holds, the way aqua-registry-g2 serves a
// registry.json, and remembers what it was asked for.
type fakeG2 struct {
	pkgs  []*lockfile.Package
	asked []string
}

func (g *fakeG2) Definitions(_ context.Context, _ *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	g.asked = append(g.asked, pkgName+"@"+version)
	return g.pkgs, nil
}

// fakeInstaller answers every registry with one package definition and remembers which
// registries it installed.
type fakeInstaller struct {
	pkgInfo   *registry.PackageInfo
	installed []string
}

func (i *fakeInstaller) InstallRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, _ string, _ *checksum.Checksums) (*registry.Config, error) {
	i.installed = append(i.installed, rgst.Name)
	return &registry.Config{PackageInfos: registry.PackageInfos{i.pkgInfo}}, nil
}

func g2Package(libc, asset, chksum string) *lockfile.Package {
	p := &lockfile.Package{
		Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64",
		Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: asset,
		Checksum: chksum, ChecksumAlgorithm: "sha256",
		Files: []*lockfile.File{{Name: "gh", Src: "bin/gh"}},
	}
	if libc != "" {
		p.Variants = map[string]string{"libc": libc}
	}
	return p
}

func privateRegistry() *aqua.Registry {
	return &aqua.Registry{Name: "private", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"}
}

func standardConfig() *aqua.Config {
	return &aqua.Config{
		Registries: aqua.Registries{
			// What "type: standard" in aqua.yaml reads as.
			"standard": {Name: "standard", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "registry.yaml"},
			"private":  privateRegistry(),
		},
	}
}

func ghEntry() *lockfile.Entry {
	return &lockfile.Entry{
		Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64",
		Checksum: "aaa", ChecksumAlgorithm: "sha256",
		Commands: []string{"gh"},
	}
}

func standardPkg() *aqua.Package {
	return &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"}
}

func locked(t *testing.T, g2 *fakeG2, rgsts *lockresolve.Registries, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error) {
	t.Helper()
	return lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry) //nolint:wrapcheck
}

// What is installed is what the registry says for the entry's environment. The lock file
// gives the checksum and picks the environment, and nothing else.
func TestLocked_standard(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{
		g2Package("glibc", "gh_linux_amd64.tar.gz", "AAA"),
		g2Package("musl", "gh_linux_amd64_musl.tar.gz", "bbb"),
	}}
	installer := &fakeInstaller{}
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml")
	entry := ghEntry()
	entry.Variants = map[string]string{"libc": "glibc"}

	got, err := locked(t, g2, rgsts, standardPkg(), entry)
	if err != nil {
		t.Fatal(err)
	}
	if got.Asset != "gh_linux_amd64.tar.gz" {
		t.Errorf("the asset is %q, want the glibc build", got.Asset)
	}
	// Case doesn't matter to the comparison, and the checksum verified against is the
	// lock file's.
	if got.Checksum != "aaa" {
		t.Errorf("the checksum is %q, want aaa", got.Checksum)
	}
	// No registry is installed for a package of the standard registry, so neither the
	// standard registry of aqua-registry v1 nor a private one is read.
	if len(installer.installed) != 0 {
		t.Errorf("installed %v, want nothing", installer.installed)
	}
}

// What the lock file records and the registry has too must agree. A difference means one
// of them changed after the entry was written.
func TestLocked_mismatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		edit  func(e *lockfile.Entry)
		isErr error
	}{
		{name: "checksum", edit: func(e *lockfile.Entry) { e.Checksum = "evil" }, isErr: lockresolve.ErrChecksumMismatch},
		{name: "no checksum", edit: func(e *lockfile.Entry) { e.Checksum = "" }, isErr: lockresolve.ErrChecksumMismatch},
		// The same value under another algorithm is a different checksum.
		{name: "algorithm", edit: func(e *lockfile.Entry) { e.ChecksumAlgorithm = "sha512" }, isErr: lockresolve.ErrChecksumMismatch},
		// A link is created for every command in the lock file, so one the registry
		// doesn't have would shadow a command of the system.
		{name: "a command the registry doesn't have", edit: func(e *lockfile.Entry) { e.Commands = []string{"gh", "git"} }, isErr: lockresolve.ErrCommandsMismatch},
		{name: "a command missing", edit: func(e *lockfile.Entry) { e.Commands = []string{"hub"} }, isErr: lockresolve.ErrCommandsMismatch},
		{
			name: "another registry",
			edit: func(e *lockfile.Entry) {
				e.Registry = &lockfile.Registry{Type: "github_content", RepoOwner: "corp", RepoName: "registry"}
			},
			isErr: lockresolve.ErrRegistryMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
			rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
			entry := ghEntry()
			tt.edit(entry)
			if _, err := locked(t, g2, rgsts, standardPkg(), entry); !errors.Is(err, tt.isErr) {
				t.Errorf("got %v, want %v", err, tt.isErr)
			}
		})
	}
}

// An entry for an environment the registry has no build for has nothing to install from.
func TestLocked_notInRegistry(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	entry := ghEntry()
	entry.OS = "darwin"
	entry.Arch = "arm64"
	if _, err := locked(t, g2, rgsts, standardPkg(), entry); !errors.Is(err, lockresolve.ErrNotInRegistry) {
		t.Errorf("got %v, want ErrNotInRegistry", err)
	}
}

// go_install downloads nothing aqua could check, so a checksum the lock file gives it is
// not used.
func TestLocked_noChecksumNeeded(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{{
		Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64",
		Type: "go_install", Path: "golang.org/x/tools/gopls",
		Files: []*lockfile.File{{Name: "gopls"}},
	}}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", Registry: "standard"}
	entry := &lockfile.Entry{
		Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64",
		Checksum: "aaa", ChecksumAlgorithm: "sha256", Commands: []string{"gopls"},
	}
	got, err := locked(t, g2, rgsts, pkg, entry)
	if err != nil {
		t.Fatal(err)
	}
	if got.Checksum != "" || got.ChecksumAlgorithm != "" {
		t.Errorf("got checksum %q (%s), want none", got.Checksum, got.ChecksumAlgorithm)
	}
}

func toolInstaller() *fakeInstaller {
	return &fakeInstaller{pkgInfo: &registry.PackageInfo{
		Name:          "corp/tool",
		Type:          "github_release",
		RepoOwner:     "corp",
		RepoName:      "tool",
		Asset:         "tool_{{.OS}}_{{.Arch}}.tar.gz",
		Format:        "tar.gz",
		SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
	}}
}

func toolEntry() *lockfile.Entry {
	return &lockfile.Entry{
		Name: "corp/tool", Version: "v1.0.0",
		Registry: &lockfile.Registry{Type: "github_content", RepoOwner: "corp", RepoName: "registry"},
		OS:       "linux", Arch: "amd64",
		Checksum: "ccc", ChecksumAlgorithm: "sha256",
		Commands: []string{"tool"},
	}
}

func toolPkg() *aqua.Package {
	return &aqua.Package{Name: "corp/tool", Version: "v1.0.0", Registry: "private"}
}

// A package from another registry is read from that registry, and only that one is
// installed. Its definition carries no checksum, so the lock file's is the one verified.
func TestLocked_otherRegistry(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{}
	installer := toolInstaller()
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml")

	for range 2 {
		got, err := locked(t, g2, rgsts, toolPkg(), toolEntry())
		if err != nil {
			t.Fatal(err)
		}
		if got.Asset != "tool_linux_amd64.tar.gz" || got.Checksum != "ccc" {
			t.Errorf("got asset %q and checksum %q", got.Asset, got.Checksum)
		}
	}
	if len(g2.asked) != 0 {
		t.Errorf("g2 was asked for %v, want nothing", g2.asked)
	}
	// Installed once, and nothing but the registry the package comes from.
	if len(installer.installed) != 1 || installer.installed[0] != "private" {
		t.Errorf("installed %v, want the private registry once", installer.installed)
	}
}

// Only the registry named "standard" that is aquaproj/aqua-registry is read from
// aqua-registry-g2. A registry named "standard" pointing elsewhere is that registry, and
// aquaproj/aqua-registry under another name is how a configuration keeps using
// aqua-registry v1.
func TestLocked_notStandard(t *testing.T) {
	t.Parallel()
	for name, rgst := range map[string]*aqua.Registry{
		"named standard, another repository": {Name: "standard", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"},
		"aqua-registry under another name":   {Name: "v1", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "registry.yaml"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
			cfg := &aqua.Config{Registries: aqua.Registries{rgst.Name: rgst}}
			rgsts := lockresolve.NewRegistries(toolInstaller(), cfg, "aqua.yaml")
			pkg := &aqua.Package{Name: "corp/tool", Version: "v1.0.0", Registry: rgst.Name}
			entry := toolEntry()
			entry.Registry = &lockfile.Registry{Type: rgst.Type, RepoOwner: rgst.RepoOwner, RepoName: rgst.RepoName}
			got, err := locked(t, g2, rgsts, pkg, entry)
			if err != nil {
				t.Fatal(err)
			}
			if len(g2.asked) != 0 {
				t.Errorf("g2 was asked for %v, want nothing", g2.asked)
			}
			if got.Asset != "tool_linux_amd64.tar.gz" {
				t.Errorf("the asset is %q, want the registry's", got.Asset)
			}
		})
	}
}
