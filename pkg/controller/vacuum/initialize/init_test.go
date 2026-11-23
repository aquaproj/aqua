package initialize_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/vacuum/initialize"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// The paths cli/cli v2.1.0 is installed at. aqua-registry v1 at aqua.yaml's ref and
// aqua-registry-g2 disagree on the repository, so they name different directories.
var (
	g1Path = filepath.Join("pkgs", "github_release", "github.com", "old", "cli", "v2.1.0", "gh.tar.gz") //nolint:gochecknoglobals
	g2Path = filepath.Join("pkgs", "github_release", "github.com", "cli", "cli", "v2.1.0", "gh.tar.gz") //nolint:gochecknoglobals
)

type fakeFinder struct{ paths []string }

func (f *fakeFinder) Finds(string, string) []string { return f.paths }

// fakeReader reads the configuration registered for each path.
type fakeReader struct{ cfgs map[string]*aqua.Config }

func (r *fakeReader) Read(_ *slog.Logger, path string, cfg *aqua.Config) error {
	*cfg = *r.cfgs[path]
	return nil
}

// fakeInstaller answers every registry with cli/cli as aqua-registry v1 defines it, and
// counts how often every registry was installed.
type fakeInstaller struct {
	installedAll int
}

func g1Info() *registry.PackageInfo {
	return &registry.PackageInfo{
		Name: "cli/cli", Type: "github_release", RepoOwner: "old", RepoName: "cli", Asset: "gh.tar.gz",
		Files: []*registry.File{{Name: "gh"}},
	}
}

func (i *fakeInstaller) InstallRegistries(context.Context, *slog.Logger, *aqua.Config, string, *checksum.Checksums) (map[string]*registry.Config, error) {
	i.installedAll++
	return map[string]*registry.Config{"standard": {PackageInfos: registry.PackageInfos{g1Info()}}}, nil
}

func (i *fakeInstaller) InstallRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return &registry.Config{PackageInfos: registry.PackageInfos{g1Info()}}, nil
}

func (i *fakeInstaller) InstallLockedRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksum) (*registry.Config, error) {
	return &registry.Config{PackageInfos: registry.PackageInfos{g1Info()}}, nil
}

// fakeG2 serves any package <owner>/<repo> as a github_release of that repository for
// linux/amd64, with the command gh.
type fakeG2 struct{}

func (fakeG2) Definitions(_ context.Context, _ *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	owner, repo, _ := strings.Cut(pkgName, "/")
	return []*lockfile.Package{{
		Name: pkgName, Version: version, OS: "linux", Arch: "amd64",
		Type: "github_release", RepoOwner: owner, RepoName: repo, Asset: "gh.tar.gz",
		Checksum: "aaa", ChecksumAlgorithm: "sha256",
		Files: []*lockfile.File{{Name: "gh"}},
	}}, nil
}

// fakeVacuum remembers the paths timestamps were created for.
type fakeVacuum struct{ created []string }

func (v *fakeVacuum) Create(pkgPath string, _ time.Time) error {
	v.created = append(v.created, pkgPath)
	return nil
}

func entry(name, goos, arch string, commands ...string) *lockfile.Entry {
	return &lockfile.Entry{
		Name: name, Version: "v2.1.0", OS: goos, Arch: arch,
		Checksum: "aaa", ChecksumAlgorithm: "sha256", Commands: commands,
	}
}

func config1(names ...string) *aqua.Config {
	cfg := &aqua.Config{Registries: aqua.Registries{
		"standard": {Name: "standard", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "registry.yaml"},
	}}
	for _, name := range names {
		cfg.Packages = append(cfg.Packages, &aqua.Package{Name: name, Version: "v2.1.0", Registry: "standard"})
	}
	return cfg
}

type harness struct {
	rootDir   string
	installer *fakeInstaller
	vacuum    *fakeVacuum
	reader    *fakeReader
	finder    *fakeFinder
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	return &harness{
		rootDir:   t.TempDir(),
		installer: &fakeInstaller{},
		vacuum:    &fakeVacuum{},
		reader:    &fakeReader{cfgs: map[string]*aqua.Config{}},
		finder:    &fakeFinder{},
	}
}

// addConfig adds a configuration file found in its own directory, with the lock file lock
// beside it: none when lock is nil, and lock written as is when it is a string.
func (h *harness) addConfig(t *testing.T, cfg *aqua.Config, lock any) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "aqua.yaml")
	h.reader.cfgs[path] = cfg
	h.finder.paths = append(h.finder.paths, path)
	lockPath := filepath.Join(dir, lockfile.FileName)
	switch lf := lock.(type) {
	case *lockfile.LockFile:
		if err := lockfile.Write(lockPath, lf); err != nil {
			t.Fatal(err)
		}
	case string:
		if err := os.WriteFile(lockPath, []byte(lf), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// install creates the directory of a package as install does.
func (h *harness) install(t *testing.T, pkgPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(h.rootDir, pkgPath), 0o755); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) init(t *testing.T) error {
	t.Helper()
	param := &config.Param{RootDir: h.rootDir}
	ctrl := initialize.New(param, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, h.vacuum, h.finder, h.reader, h.installer, fakeG2{})
	return ctrl.Init(t.Context(), slog.New(slog.DiscardHandler), param) //nolint:wrapcheck
}

// Without a lock file, every registry is read at aqua.yaml's ref, as before.
func TestController_Init_noLockFile(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.addConfig(t, config1("cli/cli"), nil)
	h.install(t, g1Path)
	if err := h.init(t); err != nil {
		t.Fatal(err)
	}
	if h.installer.installedAll != 1 {
		t.Errorf("installed every registry %d times, want once", h.installer.installedAll)
	}
	if !slices.Equal(h.vacuum.created, []string{g1Path}) {
		t.Errorf("created %v, want %v", h.vacuum.created, []string{g1Path})
	}
}

// With a lock file, the path is read from the definition install reads, aqua-registry-g2's,
// and no registry is read whole.
func TestController_Init_locked(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.addConfig(t, config1("cli/cli"), &lockfile.LockFile{Packages: []*lockfile.Entry{entry("cli/cli", "linux", "amd64", "gh")}})
	h.install(t, g2Path)
	h.install(t, g1Path)
	if err := h.init(t); err != nil {
		t.Fatal(err)
	}
	if h.installer.installedAll != 0 {
		t.Errorf("installed every registry %d times, want never", h.installer.installedAll)
	}
	if !slices.Equal(h.vacuum.created, []string{g2Path}) {
		t.Errorf("created %v, want %v", h.vacuum.created, []string{g2Path})
	}
}

// A package that isn't installed here is passed over without failing: one with an entry
// only for another environment, and one the lock file doesn't have. A package whose
// directory doesn't exist gets no timestamp.
func TestController_Init_lockedSkip(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.addConfig(t, config1("cli/cli", "other/darwin-only", "other/unlocked", "other/not-installed"), &lockfile.LockFile{Packages: []*lockfile.Entry{
		entry("cli/cli", "linux", "amd64", "gh"),
		entry("other/darwin-only", "darwin", "arm64", "gh"),
		entry("other/not-installed", "linux", "amd64", "gh"),
	}})
	h.install(t, g2Path)
	if err := h.init(t); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(h.vacuum.created, []string{g2Path}) {
		t.Errorf("created %v, want %v", h.vacuum.created, []string{g2Path})
	}
}

// A package whose definition can't be read fails the configuration, after the others are
// recorded. The next configuration isn't looked at, as when a configuration fails now.
func TestController_Init_lockedFailure(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.addConfig(t, config1("other/mismatch", "cli/cli"), &lockfile.LockFile{Packages: []*lockfile.Entry{
		entry("other/mismatch", "linux", "amd64", "git"), // the registry's command is gh
		entry("cli/cli", "linux", "amd64", "gh"),
	}})
	h.addConfig(t, config1("cli/cli"), nil)
	h.install(t, g2Path)
	h.install(t, g1Path)
	if err := h.init(t); err == nil {
		t.Fatal("an error must be returned")
	}
	if !slices.Equal(h.vacuum.created, []string{g2Path}) {
		t.Errorf("created %v, want only cli/cli of the first configuration", h.vacuum.created)
	}
	if h.installer.installedAll != 0 {
		t.Error("the second configuration must not be read")
	}
}

// A broken lock file fails rather than being taken for no lock file.
func TestController_Init_brokenLockFile(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.addConfig(t, config1("cli/cli"), "{")
	h.install(t, g1Path)
	if err := h.init(t); err == nil {
		t.Fatal("an error must be returned")
	}
	if h.installer.installedAll != 0 || len(h.vacuum.created) != 0 {
		t.Errorf("installed every registry %d times and created %v, want nothing", h.installer.installedAll, h.vacuum.created)
	}
}
