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
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// fakeG2 answers with the definitions it holds, the way aqua-registry-g2 serves a
// registry.json, and remembers what it was asked for.
type fakeG2 struct {
	pkgs  []*lockfile.Package
	asked []string
}

func (g *fakeG2) Definitions(_ context.Context, _ *slog.Logger, pkgName, version string, _ []string) (*g2.Resolved, error) {
	g.asked = append(g.asked, pkgName+"@"+version)
	return &g2.Resolved{Packages: g.pkgs}, nil
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
	pkgInfo    *registry.PackageInfo
	rootDir    string
	installed  []string
	downloaded []string
	// commits are the commits registries were downloaded or read at.
	commits []string
}

// DownloadRegistry downloads whatever is installed, as the real one does.
func (i *fakeInstaller) DownloadRegistry(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry, commit string, chks *checksum.Checksums) (*registry.Config, error) {
	i.downloaded = append(i.downloaded, rgst.Name)
	i.commits = append(i.commits, commit)
	return i.InstallRegistry(ctx, logger, rgst, "aqua.yaml", chks)
}

// InstallLockedRegistry checks the file against the recorded checksum, as the real one
// does when it downloads one.
func (i *fakeInstaller) InstallLockedRegistry(ctx context.Context, logger *slog.Logger, rgst *aqua.Registry, commit string, chk *checksum.Checksum) (*registry.Config, error) {
	i.commits = append(i.commits, commit)
	chks := checksum.New()
	chks.Set(chk.ID, chk)
	return i.InstallRegistry(ctx, logger, rgst, "", chks)
}

// testCommit is the commit a test lock file records for a registry, and the one
// fakeResolver resolves a ref to by default.
var testCommit = strings.Repeat("c", 40) //nolint:gochecknoglobals

// fakeResolver resolves every ref to commit, or fails with err, and counts the calls.
type fakeResolver struct {
	commit string
	err    error
	calls  int
}

func (r *fakeResolver) GetCommitSHA1(context.Context, string, string, string, string) (string, *github.Response, error) {
	r.calls++
	if r.err != nil {
		return "", nil, r.err
	}
	if r.commit == "" {
		return testCommit, nil, nil
	}
	return r.commit, nil, nil
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

// go_install downloads nothing aqua could check, so it takes no checksum. An entry locked
// with one is one the registry no longer describes, such as a registry changed to build
// from source what used to be downloaded, and dropping the checksum would skip the check
// the lock file recorded.
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
		Commands: []string{"gopls"},
	}
	got, err := locked(t, g2, rgsts, pkg, entry)
	if err != nil {
		t.Fatal(err)
	}
	if got.Checksum != "" || got.ChecksumAlgorithm != "" {
		t.Errorf("got checksum %q (%s), want none", got.Checksum, got.ChecksumAlgorithm)
	}

	entry.Checksum = "aaa"
	entry.ChecksumAlgorithm = "sha256"
	if _, err := locked(t, g2, rgsts, pkg, entry); !errors.Is(err, lockresolve.ErrChecksumMismatch) {
		t.Errorf("got %v, want ErrChecksumMismatch", err)
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
	key.Commit = testCommit
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
	rgsts := lockresolve.NewRecordingRegistries(installer, &fakeResolver{}, standardConfig(), "aqua.yaml", lf, nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	// Recording always downloads the registry, whatever is installed.
	if len(installer.downloaded) != 1 {
		t.Errorf("downloaded %v, want the private registry once", installer.downloaded)
	}
	// Nothing is written until it is committed, and a registry the caller leaves out
	// isn't written at all.
	if lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())) != nil {
		t.Fatal("recorded before Commit")
	}
	if rgsts.Commit(func(*lockfile.Registry) bool { return false }) || len(lf.Registries) != 0 {
		t.Fatal("a registry left out must not be written")
	}
	if !rgsts.Commit(func(*lockfile.Registry) bool { return true }) {
		t.Error("recording a checksum is a change to the lock file")
	}
	assertRecorded(t, lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())))
}

// assertRecorded checks a registry is recorded with the resolved commit and the checksum of
// the file downloaded from it.
func assertRecorded(t *testing.T, got *lockfile.Registry) {
	t.Helper()
	if got == nil || got.Commit != testCommit || got.Checksum != registryChecksum(t) || got.ChecksumAlgorithm != "sha256" {
		t.Fatalf("recorded %+v", got)
	}
}

// A recorded checksum is checked against the file on disk, and --force downloads the file
// again and replaces the checksum once it has read it.
func TestRegistries_record_recorded(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	installer := toolInstaller()
	installer.rootDir = rootDir
	if err := writeRegistryFile(privateRegistry(), rootDir, "aqua.yaml"); err != nil {
		t.Fatal(err)
	}
	rgsts := lockresolve.NewRecordingRegistries(installer, &fakeResolver{}, standardConfig(), "aqua.yaml", lockFileWithRegistry(privateRegistry(), "0000"), nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err == nil {
		t.Fatal("an error must be returned")
	}

	lf := lockFileWithRegistry(privateRegistry(), "0000")
	rgsts = lockresolve.NewRecordingRegistries(installer, &fakeResolver{}, standardConfig(), "aqua.yaml", lf, forceAll)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	rgsts.Commit(func(*lockfile.Registry) bool { return true })
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
	rgsts := lockresolve.NewRecordingRegistries(installer, &fakeResolver{}, standardConfig(), "aqua.yaml", lf, nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	rgsts.Commit(func(*lockfile.Registry) bool { return true })
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
		"another file of aqua-registry":      {Name: "standard", Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry", Ref: "v4.0.0", Path: "other.yaml"},
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

// A lock file entry whose linked libc isn't the registry's describes a build the registry
// doesn't have: Find would pick it as the static fallback on a machine it may not run on.
func TestLocked_linkedLibc(t *testing.T) {
	t.Parallel()
	def := g2Package("musl", "gh_musl.tar.gz", "aaa")
	def.LinkedLibc = "musl"
	g2 := &fakeG2{pkgs: []*lockfile.Package{def}}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", nil)
	entry := ghEntry()
	entry.Variants = map[string]string{"libc": "musl"}
	entry.LinkedLibc = "static"
	if _, err := locked(t, g2, rgsts, standardPkg(), entry); !errors.Is(err, lockresolve.ErrNotInRegistry) {
		t.Errorf("got %v, want ErrNotInRegistry", err)
	}
}

// Without aqua-registry-g2 a package of the standard registry can't be read, and that is
// an error rather than a panic.
func TestLocked_noG2(t *testing.T) {
	t.Parallel()
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", nil)
	if _, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), nil, rgsts, standardPkg(), ghEntry()); err == nil {
		t.Fatal("an error must be returned")
	}
}

// failingInstaller fails to download any registry, as offline.
type failingInstaller struct{}

func (failingInstaller) InstallRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return nil, errors.New("offline")
}

func (failingInstaller) InstallLockedRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksum) (*registry.Config, error) {
	return nil, errors.New("offline")
}

func (failingInstaller) DownloadRegistry(context.Context, *slog.Logger, *aqua.Registry, string, *checksum.Checksums) (*registry.Config, error) {
	return nil, errors.New("offline")
}

// The registry file on disk is shared by every configuration using the same registry at the
// same ref. A download that fails while recording leaves it as it was.
func TestRegistries_record_failedDownload(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	if err := writeRegistryFile(privateRegistry(), rootDir, "aqua.yaml"); err != nil {
		t.Fatal(err)
	}
	rgsts := lockresolve.NewRecordingRegistries(failingInstaller{}, &fakeResolver{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{}, forceAll)
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
}

func forceAll(*lockfile.Registry) bool { return true }

// A registry the lock file has a commit for is downloaded from that commit, and the ref
// isn't resolved: a ref that has moved on doesn't change the lock file without --force.
func TestRegistries_record_keepsCommit(t *testing.T) {
	t.Parallel()
	installer := toolInstaller()
	resolver := &fakeResolver{commit: strings.Repeat("d", 40)}
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	rgsts := lockresolve.NewRecordingRegistries(installer, resolver, standardConfig(), "aqua.yaml", lf, nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 0 {
		t.Errorf("resolved the ref %d times, want none", resolver.calls)
	}
	if len(installer.commits) != 1 || installer.commits[0] != testCommit {
		t.Errorf("downloaded at %v, want the recorded commit", installer.commits)
	}
	if rgsts.Commit(func(*lockfile.Registry) bool { return true }) {
		t.Error("checking a recorded registry must not change the lock file")
	}
}

// A record without the commit, from a lock file written before the commit was recorded,
// counts as none: the ref is resolved and the commit is written, even though the checksum
// is the same.
func TestRegistries_record_withoutCommit(t *testing.T) {
	t.Parallel()
	resolver := &fakeResolver{}
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	lf.Registries[0].Commit = ""
	rgsts := lockresolve.NewRecordingRegistries(toolInstaller(), resolver, standardConfig(), "aqua.yaml", lf, nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	if resolver.calls != 1 {
		t.Errorf("resolved the ref %d times, want once", resolver.calls)
	}
	if !rgsts.Commit(func(*lockfile.Registry) bool { return true }) {
		t.Fatal("writing the commit is a change to the lock file")
	}
	if got := lf.FindRegistry(lockresolve.LockFileRegistry(privateRegistry())); got.Commit != testCommit || len(lf.Registries) != 1 {
		t.Errorf("recorded %+v of %d", got, len(lf.Registries))
	}
}

// A ref that can't be resolved fails the registry, and nothing is recorded for it.
func TestRegistries_record_unresolvableRef(t *testing.T) {
	t.Parallel()
	installer := toolInstaller()
	lf := &lockfile.LockFile{}
	rgsts := lockresolve.NewRecordingRegistries(installer, &fakeResolver{err: errors.New("not found")}, standardConfig(), "aqua.yaml", lf, nil)
	if _, err := rgsts.Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err == nil {
		t.Fatal("an error must be returned")
	}
	if len(installer.downloaded) != 0 {
		t.Errorf("downloaded %v, want nothing", installer.downloaded)
	}
	if rgsts.Commit(func(*lockfile.Registry) bool { return true }) || len(lf.Registries) != 0 {
		t.Error("nothing must be recorded for a registry whose ref can't be resolved")
	}
}

// Install and exec read a registry at the recorded commit, and refuse a record without
// the commit, which the lock update command records again.
func TestLocked_registryCommit(t *testing.T) {
	t.Parallel()
	installer := toolInstaller()
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	if _, err := lockresolve.NewRegistries(installer, standardConfig(), "aqua.yaml", lf).Get(t.Context(), slog.New(slog.DiscardHandler), "private"); err != nil {
		t.Fatal(err)
	}
	if len(installer.commits) != 1 || installer.commits[0] != testCommit {
		t.Errorf("read at %v, want the recorded commit", installer.commits)
	}

	lf.Registries[0].Commit = ""
	_, err := lockresolve.NewRegistries(toolInstaller(), standardConfig(), "aqua.yaml", lf).Get(t.Context(), slog.New(slog.DiscardHandler), "private")
	if !errors.Is(err, lockresolve.ErrRegistryNotInLockFile) {
		t.Errorf("got %v, want ErrRegistryNotInLockFile", err)
	}
}

// A registry that can't be read at the recorded commit, which may have become unavailable,
// points to recording the commit again.
func TestLocked_registryCommitUnavailable(t *testing.T) {
	t.Parallel()
	lf := lockFileWithRegistry(privateRegistry(), registryChecksum(t))
	_, err := lockresolve.NewRegistries(failingInstaller{}, standardConfig(), "aqua.yaml", lf).Get(t.Context(), slog.New(slog.DiscardHandler), "private")
	if err == nil {
		t.Fatal("an error must be returned")
	}
	// The hint is logged with the error, as aqua reports one.
	buf := &strings.Builder{}
	slogerr.WithError(slog.New(slog.NewTextHandler(buf, nil)), err).Error("install the registry")
	if !strings.Contains(buf.String(), `aqua lock update --force`) || !strings.Contains(buf.String(), testCommit) {
		t.Errorf("logged %q, want the commit and a hint pointing to aqua lock update --force", buf.String())
	}
}

// templateG2 answers the way aqua-registry-g2 does for a package it serves as a template:
// entries expanded for the environments asked for, with no checksum, and the definition
// they were expanded from.
type templateG2 struct {
	envs [][]string
}

func (g *templateG2) Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*g2.Resolved, error) {
	return g.Definitions(ctx, logger, pkgName, version, supportedEnvs)
}

func (g *templateG2) Definitions(_ context.Context, _ *slog.Logger, pkgName, version string, supportedEnvs []string) (*g2.Resolved, error) {
	g.envs = append(g.envs, supportedEnvs)
	pkg := g2Package("", "gh_linux_amd64.tar.gz", "")
	pkg.ChecksumAlgorithm = ""
	pkg.Name = pkgName
	pkg.Version = version
	return &g2.Resolved{
		Packages: []*lockfile.Package{pkg},
		PkgInfo:  &registry.PackageInfo{Name: pkgName, Type: "github_release"},
	}, nil
}

// A package served as a template is expanded for the environments the configuration locks,
// and comes back with what its checksums are fetched from, as one from another registry does.
func TestResolve_template(t *testing.T) {
	t.Parallel()
	cfg := standardConfig()
	cfg.Checksum = &aqua.Checksum{SupportedEnvs: []string{"linux/amd64"}}
	g := &templateG2{}
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, cfg, "aqua.yaml", &lockfile.LockFile{})

	defs, err := lockresolve.Resolve(t.Context(), slog.New(slog.DiscardHandler), g, rgsts, standardPkg())
	if err != nil {
		t.Fatal(err)
	}
	if defs.PkgInfo == nil {
		t.Fatal("the definition the checksums are fetched from must be returned")
	}
	if diff := cmp.Diff([]string{"linux/amd64"}, defs.SupportedEnvs); diff != "" {
		t.Errorf("supported envs are wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([][]string{{"linux/amd64"}}, g.envs); diff != "" {
		t.Errorf("the template was expanded for the wrong environments (-want +got):\n%s", diff)
	}
}

// An entry of a template is installed with the lock file's checksum, the registry having none.
func TestLocked_template(t *testing.T) {
	t.Parallel()
	rgsts := lockresolve.NewRegistries(&fakeInstaller{}, standardConfig(), "aqua.yaml", &lockfile.LockFile{})
	got, err := lockresolve.Locked(t.Context(), slog.New(slog.DiscardHandler), &templateG2{}, rgsts, standardPkg(), ghEntry())
	if err != nil {
		t.Fatal(err)
	}
	if got.Checksum != "aaa" || got.ChecksumAlgorithm != "sha256" {
		t.Errorf("the checksum is %s:%s, want the lock file's", got.ChecksumAlgorithm, got.Checksum)
	}
}
