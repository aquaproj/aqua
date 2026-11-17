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

func (g *fakeG2) Resolve(_ context.Context, _ *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	g.asked = append(g.asked, pkgName+"@"+version)
	return g.pkgs, nil
}

// fakeInstaller answers every registry with one package definition, and remembers which
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
	}
	if libc != "" {
		p.Variants = map[string]string{"libc": libc}
	}
	return p
}

func standardConfig() *aqua.Config {
	return &aqua.Config{
		Registries: aqua.Registries{
			// What "type: standard" in aqua.yaml reads as.
			"standard": {Name: "standard", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "registry.yaml"},
			"private":  {Name: "private", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"},
		},
	}
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
	pkg := &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"}
	entry := &lockfile.Entry{
		Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64",
		Variants: map[string]string{"libc": "glibc"},
		Checksum: "aaa", ChecksumAlgorithm: "sha256",
	}

	got, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry)
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

// A lock file whose checksum doesn't match the registry's has been changed, or the
// registry has. Installing either would install something nobody locked.
func TestLocked_checksumMismatch(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"}
	for _, chksum := range []string{"evil", ""} {
		entry := &lockfile.Entry{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: chksum}
		_, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry)
		if !errors.Is(err, lockresolve.ErrChecksumMismatch) {
			t.Errorf("checksum %q: got %v, want ErrChecksumMismatch", chksum, err)
		}
	}
}

// An entry for an environment the registry has no build for has nothing to install from.
func TestLocked_notInRegistry(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"}
	entry := &lockfile.Entry{Name: "cli/cli", Version: "v2.1.0", OS: "darwin", Arch: "arm64", Checksum: "aaa"}
	if _, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry); !errors.Is(err, lockresolve.ErrNotInRegistry) {
		t.Errorf("got %v, want ErrNotInRegistry", err)
	}
}

// A package from another registry is read from that registry, and only that one is
// installed. Its definition carries no checksum, so the lock file's is the one verified.
func TestLocked_otherRegistry(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{}
	installer := &fakeInstaller{pkgInfo: &registry.PackageInfo{
		Name:          "corp/tool",
		Type:          "github_release",
		RepoOwner:     "corp",
		RepoName:      "tool",
		Asset:         "tool_{{.OS}}_{{.Arch}}.tar.gz",
		Format:        "tar.gz",
		SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
	}}
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "corp/tool", Version: "v1.0.0", Registry: "private"}
	entry := &lockfile.Entry{Name: "corp/tool", Version: "v1.0.0", OS: "linux", Arch: "amd64", Checksum: "ccc", ChecksumAlgorithm: "sha256"}

	for range 2 {
		got, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry)
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
			installer := &fakeInstaller{pkgInfo: &registry.PackageInfo{
				Name: "cli/cli", Type: "github_release", RepoOwner: "cli", RepoName: "cli",
				Asset: "gh_{{.OS}}_{{.Arch}}.tar.gz", SupportedEnvs: registry.SupportedEnvs{"linux/amd64"},
			}}
			cfg := &aqua.Config{Registries: aqua.Registries{rgst.Name: rgst}}
			rgsts := lockresolve.NewRegistries(installer, cfg, "aqua.yaml")
			pkg := &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: rgst.Name}
			entry := &lockfile.Entry{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "ccc", ChecksumAlgorithm: "sha256"}
			got, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry)
			if err != nil {
				t.Fatal(err)
			}
			if len(g2.asked) != 0 {
				t.Errorf("g2 was asked for %v, want nothing", g2.asked)
			}
			if got.Asset != "gh_linux_amd64.tar.gz" {
				t.Errorf("the asset is %q, want the registry's", got.Asset)
			}
		})
	}
}

// The same value under another algorithm is a different checksum.
func TestLocked_algorithmMismatch(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{g2Package("", "gh.tar.gz", "aaa")}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"}
	entry := &lockfile.Entry{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Checksum: "aaa", ChecksumAlgorithm: "sha512"}
	if _, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry); !errors.Is(err, lockresolve.ErrChecksumMismatch) {
		t.Errorf("got %v, want ErrChecksumMismatch", err)
	}
}

// go_install downloads nothing aqua could check, so a checksum the lock file gives it is
// not used.
func TestLocked_noChecksumNeeded(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{pkgs: []*lockfile.Package{{
		Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64",
		Type: "go_install", Path: "golang.org/x/tools/gopls",
	}}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml")
	pkg := &aqua.Package{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", Registry: "standard"}
	entry := &lockfile.Entry{Name: "golang.org/x/tools/gopls", Version: "v0.14.0", OS: "linux", Arch: "amd64", Checksum: "aaa", ChecksumAlgorithm: "sha256"}
	got, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), g2, rgsts, pkg, entry)
	if err != nil {
		t.Fatal(err)
	}
	if got.Checksum != "" || got.ChecksumAlgorithm != "" {
		t.Errorf("got checksum %q (%s), want none", got.Checksum, got.ChecksumAlgorithm)
	}
}
