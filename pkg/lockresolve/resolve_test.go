package lockresolve_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

// registryFile is the content every fake registry file has.
const registryFile = "packages: []\n"

// registryChecksum is the sha256 of registryFile.
func registryChecksum(t *testing.T) string {
	t.Helper()
	sum, err := checksum.CalculateReader(strings.NewReader(registryFile), "sha256")
	if err != nil {
		t.Fatal(err)
	}
	return strings.ToLower(sum)
}

// fakeInstaller answers every registry with one package definition and remembers which
// registries it installed. It checks the file against the checksums the way the real
// installer does when it downloads one, and writes it where the real one would.
type fakeInstaller struct {
	pkgInfo   *registry.PackageInfo
	rootDir   string
	installed []string
}

func (i *fakeInstaller) InstallRegistry(_ context.Context, _ *slog.Logger, rgst *aqua.Registry, cfgFilePath string, chks *checksum.Checksums) (*registry.Config, error) {
	i.installed = append(i.installed, rgst.Name)
	if rgst.Type == aqua.RegistryTypeGitHubContent {
		if err := checksum.CheckRegistry(rgst, chks, []byte(registryFile)); err != nil {
			return nil, err //nolint:wrapcheck
		}
		if i.rootDir != "" {
			if err := writeRegistryFile(rgst, i.rootDir, cfgFilePath); err != nil {
				return nil, err
			}
		}
	}
	return &registry.Config{PackageInfos: registry.PackageInfos{i.pkgInfo}}, nil
}

func writeRegistryFile(rgst *aqua.Registry, rootDir, cfgFilePath string) error {
	p, err := rgst.FilePath(rootDir, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err //nolint:wrapcheck
	}
	return os.WriteFile(p, []byte(registryFile), 0o600) //nolint:wrapcheck
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
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml", &lockfile.LockFile{})
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
			rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{})
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
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{})
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
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{})
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

// lockFileWithRegistry is a lock file recording chksum for the file of rgst.
func lockFileWithRegistry(rgst *aqua.Registry, chksum string) *lockfile.LockFile {
	key := lockresolve.LockFileRegistry(rgst)
	key.Checksum = chksum
	key.ChecksumAlgorithm = "sha256"
	return &lockfile.LockFile{Registries: []*lockfile.Registry{key}}
}

// A package from another registry is read from that registry, and only that one is
// installed. Its definition carries no checksum, so the lock file's is the one verified.
func TestLocked_otherRegistry(t *testing.T) {
	t.Parallel()
	g2 := &fakeG2{}
	installer := toolInstaller()
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	rgsts := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml", lf)

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

// The registry decides where a package is downloaded from, so a registry file is read only
// when its checksum is the one the lock file recorded, and the lock file has to have one.
func TestLocked_registryChecksum(t *testing.T) {
	t.Parallel()
	for name, lf := range map[string]*lockfile.LockFile{
		"another checksum": lockFileWithRegistry(privateRegistry(), "0000"),
		"no checksum":      {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rgsts := lockresolve.NewRegistries(toolInstaller(), standardConfig(), "aqua.yaml", lf)
			if _, err := locked(t, &fakeG2{}, rgsts, toolPkg(), toolEntry()); err == nil {
				t.Fatal("an error must be returned")
			}
		})
	}
}

// The lock update command records the checksum of a registry file, and checks one it
// already has rather than replacing it.
func TestRegistries_record(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	lf := &lockfile.LockFile{}
	installer := toolInstaller()
	installer.rootDir = rootDir
	rgsts := lockresolve.NewRecordingRegistries(installer, standardConfig(), "aqua.yaml", lf, rootDir, false)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	got := lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry()))
	if got == nil || got.Checksum != registryChecksum(t) || got.ChecksumAlgorithm != "sha256" {
		t.Fatalf("recorded %+v", got)
	}
	if !rgsts.Changed() {
		t.Error("recording a checksum is a change to the lock file")
	}

	// The file on disk no longer matches what was recorded.
	rgsts = lockresolve.NewRecordingRegistries(installer, standardConfig(), "aqua.yaml", lockFileWithRegistry(privateRegistry(), "0000"), rootDir, false)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err == nil {
		t.Fatal("an error must be returned")
	}

	// --force downloads the file again and replaces the checksum once it has read it.
	lf = lockFileWithRegistry(privateRegistry(), "0000")
	rgsts = lockresolve.NewRecordingRegistries(installer, standardConfig(), "aqua.yaml", lf, rootDir, true)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	if got := lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())); got.Checksum != registryChecksum(t) || len(lf.Registries) != 1 {
		t.Errorf("recorded %+v of %d", got, len(lf.Registries))
	}
}

// A checksum is recorded from a file just downloaded, not from one already on disk, which
// may be stale or edited.
func TestRegistries_recordDownloaded(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	p, err := privateRegistry().FilePath(rootDir, "aqua.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("edited\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	lf := &lockfile.LockFile{}
	installer := toolInstaller()
	installer.rootDir = rootDir
	rgsts := lockresolve.NewRecordingRegistries(installer, standardConfig(), "aqua.yaml", lf, rootDir, false)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	if got := lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())); got == nil || got.Checksum != registryChecksum(t) {
		t.Errorf("recorded %+v, want the checksum of the downloaded file", got)
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
			rgsts := lockresolve.NewRegistries(toolInstaller(), cfg, "aqua.yaml", lockFileWithRegistry(rgst, registryChecksum(t)))
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

// A registry checksum of an algorithm that isn't collision resistant can't stand for the
// registry file.
func TestLocked_weakRegistryChecksum(t *testing.T) {
	t.Parallel()
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	lf.Registries[0].ChecksumAlgorithm = "md5"
	rgsts := lockresolve.NewRegistries(toolInstaller(), standardConfig(), "aqua.yaml", lf)
	if _, err := locked(t, &fakeG2{}, rgsts, toolPkg(), toolEntry()); !errors.Is(err, lockresolve.ErrWeakRegistryChecksum) {
		t.Errorf("got %v, want ErrWeakRegistryChecksum", err)
	}
}

// failingInstaller fails to download any registry, as offline.
type failingInstaller struct{}

func (failingInstaller) InstallRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return nil, errors.New("offline")
}

// The registry file on disk is shared by every configuration using the same registry at the
// same ref. Recording downloads it again, and a download that fails puts the old copy back.
func TestRegistries_record_failedDownload(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	if err := writeRegistryFile(privateRegistry(), rootDir, "aqua.yaml"); err != nil {
		t.Fatal(err)
	}
	rgsts := lockresolve.NewRecordingRegistries(failingInstaller{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{}, rootDir, true)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err == nil {
		t.Fatal("an error must be returned")
	}
	p, err := privateRegistry().FilePath(rootDir, "aqua.yaml")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("the registry file must be put back: %v", err)
	}
	if string(b) != registryFile {
		t.Errorf("the registry file is %q", b)
	}
	if _, err := os.Stat(p + ".aqua-lock-update"); !os.IsNotExist(err) {
		t.Error("nothing must be left aside")
	}
}
