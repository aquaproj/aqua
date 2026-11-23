package remove_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/remove"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

var errNoConfig = errors.New("no configuration file")

type fakeFinder struct{ path string }

func (f *fakeFinder) Find(string, string, ...string) (string, error) {
	if f.path == "" {
		return "", errNoConfig
	}
	return f.path, nil
}

type fakeReader struct{ cfg *aqua.Config }

func (r *fakeReader) Read(_ *slog.Logger, _ string, cfg *aqua.Config) error {
	*cfg = *r.cfg
	return nil
}

// ghInfo is cli/cli as the registry defines it.
func ghInfo() *registry.PackageInfo {
	return &registry.PackageInfo{
		Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz",
		Files: []*registry.File{{Name: "gh"}},
	}
}

// fakeInstaller answers every registry with cli/cli and counts how often every registry
// was installed.
type fakeInstaller struct {
	installedAll int
}

func (i *fakeInstaller) InstallRegistries(context.Context, *slog.Logger, *aqua.Config, string, *checksum.Checksums) (map[string]*registry.Config, error) {
	i.installedAll++
	return map[string]*registry.Config{"standard": {PackageInfos: registry.PackageInfos{ghInfo()}}}, nil
}

func (i *fakeInstaller) InstallRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return &registry.Config{PackageInfos: registry.PackageInfos{ghInfo()}}, nil
}

func (i *fakeInstaller) InstallLockedRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksum) (*registry.Config, error) {
	return &registry.Config{PackageInfos: registry.PackageInfos{ghInfo()}}, nil
}

// fakeG2 serves cli/cli's definitions for linux/amd64 and darwin/arm64, and remembers what
// it was asked for.
type fakeG2 struct {
	asked []string
}

func (g *fakeG2) Definitions(_ context.Context, _ *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	g.asked = append(g.asked, pkgName+"@"+version)
	pkgs := make([]*lockfile.Package, 0, 2)
	for _, env := range [][2]string{{"linux", "amd64"}, {"darwin", "arm64"}} {
		pkgs = append(pkgs, &lockfile.Package{
			Name: pkgName, Version: version, OS: env[0], Arch: env[1],
			Type: "github_release", RepoOwner: "cli", RepoName: "cli", Asset: "gh.tar.gz",
			Checksum: "aaa", ChecksumAlgorithm: "sha256",
			Files: []*lockfile.File{{Name: "gh"}},
		})
	}
	return pkgs, nil
}

// fakeWhich finds gh as cli/cli, and nothing else.
type fakeWhich struct{}

func (fakeWhich) Which(_ context.Context, _ *slog.Logger, _ *config.Param, exeName string) (*which.FindResult, error) {
	if exeName != "gh" {
		return nil, which.ErrCommandIsNotFound
	}
	return &which.FindResult{Package: &config.Package{
		Package:     &aqua.Package{Name: "cli/cli", Version: "v2.1.0", Registry: "standard"},
		PackageInfo: ghInfo(),
	}}, nil
}

// abortingFinder aborts choosing packages interactively.
type abortingFinder struct{}

func (abortingFinder) FindMulti([]*fuzzyfinder.Item, bool) ([]int, error) {
	return nil, fuzzyfinder.ErrAbort
}

type fakeVacuum struct{}

func (fakeVacuum) Remove(string) error { return nil }

// ghEntry is cli/cli's entry for linux/amd64.
func ghEntry(version string) *lockfile.Entry {
	return &lockfile.Entry{
		Name: "cli/cli", Version: version, OS: "linux", Arch: "amd64",
		Checksum: "aaa", ChecksumAlgorithm: "sha256", Commands: []string{"gh"},
	}
}

func ghConfig(versions ...string) *aqua.Config {
	cfg := &aqua.Config{Registries: aqua.Registries{
		"standard": {Name: "standard", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "registry.yaml"},
	}}
	for _, v := range versions {
		cfg.Packages = append(cfg.Packages, &aqua.Package{Name: "cli/cli", Version: v, Registry: "standard"})
	}
	return cfg
}

type harness struct {
	rootDir   string
	installer *fakeInstaller
	g2        *fakeG2
	ctrl      *remove.Controller
}

// newHarness installs cli/cli at two versions and links gh, writes lf beside aqua.yaml
// unless it is nil, and creates a controller for rt. With no cfg, no configuration file
// is found.
func newHarness(t *testing.T, cfg *aqua.Config, lf *lockfile.LockFile, rt *runtime.Runtime) *harness {
	t.Helper()
	rootDir := t.TempDir()
	for _, p := range []string{
		filepath.Join(rootDir, "pkgs", "github_release", "github.com", "cli", "cli", "v2.0.0", "gh.tar.gz", "gh"),
		filepath.Join(rootDir, "pkgs", "github_release", "github.com", "cli", "cli", "v2.1.0", "gh.tar.gz", "gh"),
		filepath.Join(rootDir, "bin", "gh"),
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	finder := &fakeFinder{}
	if cfg != nil {
		dir := t.TempDir()
		finder.path = filepath.Join(dir, "aqua.yaml")
		if lf != nil {
			if err := lockfile.Write(filepath.Join(dir, lockfile.FileName), lf); err != nil {
				t.Fatal(err)
			}
		}
	} else {
		cfg = &aqua.Config{}
	}
	h := &harness{rootDir: rootDir, installer: &fakeInstaller{}, g2: &fakeG2{}}
	h.ctrl = remove.New(&config.Param{RootDir: rootDir}, &config.RemoveMode{Link: true, Package: true}, rt,
		finder, &fakeReader{cfg: cfg}, h.installer, abortingFinder{}, fakeWhich{}, fakeVacuum{}, h.g2)
	return h
}

func linux() *runtime.Runtime { return &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"} }

func lockFile(entries ...*lockfile.Entry) *lockfile.LockFile {
	lf := lockfile.New()
	lf.Packages = entries
	return lf
}

func (h *harness) remove(t *testing.T, param *config.Param) error {
	t.Helper()
	param.RootDir = h.rootDir
	return h.ctrl.Remove(t.Context(), slog.New(slog.DiscardHandler), param) //nolint:wrapcheck
}

func (h *harness) assertRemoved(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(h.rootDir, "pkgs", "github_release", "github.com", "cli", "cli")); !os.IsNotExist(err) {
		t.Error("every version of cli/cli must be removed")
	}
	if _, err := os.Stat(filepath.Join(h.rootDir, "bin", "gh")); !os.IsNotExist(err) {
		t.Error("the link of gh must be removed")
	}
}

func (h *harness) assertKept(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(h.rootDir, "pkgs", "github_release", "github.com", "cli", "cli", "v2.1.0")); err != nil {
		t.Errorf("cli/cli must be kept: %v", err)
	}
}

// A command is removed through which alone: no registry is installed, and no configuration
// file is needed.
func TestController_Remove_commandOnly(t *testing.T) {
	t.Parallel()
	for name, cfg := range map[string]*aqua.Config{"no configuration file": nil, "with a lock file": ghConfig("v2.1.0")} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t, cfg, lockFile(ghEntry("v2.1.0")), linux())
			if err := h.remove(t, &config.Param{Args: []string{"gh"}}); err != nil {
				t.Fatal(err)
			}
			h.assertRemoved(t)
			if h.installer.installedAll != 0 {
				t.Errorf("every registry was installed %d times, want none", h.installer.installedAll)
			}
		})
	}
}

// A package the lock file has is resolved as which resolves one, from aqua-registry-g2
// for the standard registry, without installing every registry. Every version of it is
// removed, its path being the repository's directory.
func TestController_Remove_lockedPackage(t *testing.T) {
	t.Parallel()
	h := newHarness(t, ghConfig("v2.1.0"), lockFile(ghEntry("v2.1.0")), linux())
	if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}}); err != nil {
		t.Fatal(err)
	}
	h.assertRemoved(t)
	if h.installer.installedAll != 0 {
		t.Errorf("every registry was installed %d times, want none", h.installer.installedAll)
	}
	if len(h.g2.asked) != 1 || h.g2.asked[0] != "cli/cli@v2.1.0" {
		t.Errorf("asked aqua-registry-g2 for %v, want cli/cli@v2.1.0", h.g2.asked)
	}
}

// A package the lock file has entries of for other environments alone is resolved through
// one of them.
func TestController_Remove_lockedPackageOtherEnvironment(t *testing.T) {
	t.Parallel()
	// The lock file has an entry for linux/amd64 alone, and this is darwin/arm64.
	h := newHarness(t, ghConfig("v2.1.0"), lockFile(ghEntry("v2.1.0")), &runtime.Runtime{GOOS: "darwin", GOARCH: "arm64"})
	if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}}); err != nil {
		t.Fatal(err)
	}
	h.assertRemoved(t)
	if h.installer.installedAll != 0 {
		t.Errorf("every registry was installed %d times, want none", h.installer.installedAll)
	}
}

// Every version aqua.yaml declares the package at is resolved.
func TestController_Remove_lockedPackageVersions(t *testing.T) {
	t.Parallel()
	h := newHarness(t, ghConfig("v2.0.0", "v2.1.0"), lockFile(ghEntry("v2.0.0"), ghEntry("v2.1.0")), linux())
	if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}}); err != nil {
		t.Fatal(err)
	}
	h.assertRemoved(t)
	if len(h.g2.asked) != 2 {
		t.Errorf("asked aqua-registry-g2 for %v, want both versions", h.g2.asked)
	}
}

// A package that isn't in the lock file, or isn't declared in aqua.yaml, and a package
// without a lock file are found among every registry, as before.
func TestController_Remove_unlockedPackage(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		cfg *aqua.Config
		lf  *lockfile.LockFile
	}{
		"not in the lock file": {cfg: ghConfig("v2.1.0"), lf: lockFile()},
		"not in aqua.yaml":     {cfg: ghConfig(), lf: lockFile(ghEntry("v2.1.0"))},
		"no lock file":         {cfg: ghConfig("v2.1.0")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t, tt.cfg, tt.lf, linux())
			if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}}); err != nil {
				t.Fatal(err)
			}
			h.assertRemoved(t)
			if h.installer.installedAll != 1 {
				t.Errorf("every registry was installed %d times, want once", h.installer.installedAll)
			}
		})
	}
}

// A lock file that disagrees with the registry fails, as it does for which, rather than
// falling back to every registry.
func TestController_Remove_lockMismatch(t *testing.T) {
	t.Parallel()
	entry := ghEntry("v2.1.0")
	entry.Commands = []string{"other"}
	h := newHarness(t, ghConfig("v2.1.0"), lockFile(entry), linux())
	if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}}); err == nil {
		t.Fatal("an error must be returned")
	}
	h.assertKept(t)
	if h.installer.installedAll != 0 {
		t.Errorf("every registry was installed %d times, want none", h.installer.installedAll)
	}
}

// Choosing packages interactively installs every registry, and a package name given with
// it is left alone, as before.
func TestController_Remove_interactive(t *testing.T) {
	t.Parallel()
	h := newHarness(t, ghConfig("v2.1.0"), lockFile(ghEntry("v2.1.0")), linux())
	if err := h.remove(t, &config.Param{Args: []string{"cli/cli"}, SelectVersion: true}); err != nil {
		t.Fatal(err)
	}
	h.assertKept(t)
	if h.installer.installedAll != 1 {
		t.Errorf("every registry was installed %d times, want once", h.installer.installedAll)
	}
}

// A command and a package name together: the command through which, the package through
// the lock file.
func TestController_Remove_mixed(t *testing.T) {
	t.Parallel()
	h := newHarness(t, ghConfig("v2.1.0"), lockFile(ghEntry("v2.1.0")), linux())
	if err := h.remove(t, &config.Param{Args: []string{"gh", "cli/cli"}}); err != nil {
		t.Fatal(err)
	}
	h.assertRemoved(t)
	if h.installer.installedAll != 0 {
		t.Errorf("every registry was installed %d times, want none", h.installer.installedAll)
	}
}
