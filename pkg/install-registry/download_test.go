package registry_test

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
)

const servedRegistry = `packages:
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
`

// servingDownloader serves the registry file, and counts how often it was asked. With
// offline, it fails instead, to show a registry was read from disk.
type servingDownloader struct {
	content string
	offline bool
	calls   int
}

func (d *servingDownloader) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, _ *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	d.calls++
	if d.offline {
		return nil, errors.New("offline")
	}
	return &domain.GitHubContentFile{String: d.content}, nil
}

func newInstaller(rootDir string, d *servingDownloader) *registry.Installer {
	return registry.New(&config.Param{RootDir: rootDir, MaxParallelism: 5}, d, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
}

func corpRegistry() *aqua.Registry {
	return &aqua.Registry{Name: "corp", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"}
}

func sha256Of(t *testing.T, content string) string {
	t.Helper()
	sum, err := checksum.CalculateReader(strings.NewReader(content), "sha256")
	if err != nil {
		t.Fatal(err)
	}
	return strings.ToLower(sum)
}

func lockedPath(t *testing.T, rootDir string, rgst *aqua.Registry, sum string) string {
	t.Helper()
	p, err := rgst.LockedFilePath(rootDir, "sha256", sum)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func recorded(rgst *aqua.Registry, sum string) *checksum.Checksum {
	id := checksum.RegistryID(rgst)
	return &checksum.Checksum{ID: id, Checksum: sum, Algorithm: "sha256"}
}

// DownloadRegistry downloads the registry even when a file is installed at the ref's path,
// records its checksum, and keeps it under that checksum, with the JSON made from it. The
// file at the ref's path is left alone: it may be what another lock file recorded.
func TestInstaller_DownloadRegistry(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	cfgFilePath := filepath.Join(rootDir, "aqua.yaml")
	rgst := corpRegistry()
	old, err := rgst.FilePath(rootDir, cfgFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(old), 0o755); err != nil {
		t.Fatal(err)
	}
	writeStale(t, old)

	d := &servingDownloader{content: servedRegistry}
	chks := checksum.New()
	got, err := newInstaller(rootDir, d).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, cfgFilePath, chks)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PackageInfos) != 1 || got.PackageInfos[0].RepoName != "ci-info" {
		t.Errorf("got %+v, want the served registry", got.PackageInfos)
	}
	sum := sha256Of(t, servedRegistry)
	checkRecorded(t, chks, rgst, sum)
	p := lockedPath(t, rootDir, rgst, sum)
	assertContent(t, p, servedRegistry)
	if _, err := os.Stat(p + ".json"); err != nil {
		t.Errorf("the JSON made from the registry must be kept beside it: %v", err)
	}
	assertContent(t, old, "edited\n")
	assertNoTemporaryFiles(t, rootDir)
	assertRegistryFilesArePrivate(t, filepath.Join(rootDir, "registries"))

	assertReadsOffline(t, rootDir, rgst, sum)
}

// assertReadsOffline checks InstallLockedRegistry reads the kept registry without
// downloading it again.
func assertReadsOffline(t *testing.T, rootDir string, rgst *aqua.Registry, sum string) {
	t.Helper()
	offline := &servingDownloader{offline: true}
	installed, err := newInstaller(rootDir, offline).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, recorded(rgst, sum))
	if err != nil {
		t.Fatal(err)
	}
	if len(installed.PackageInfos) != 1 || installed.PackageInfos[0].RepoName != "ci-info" {
		t.Errorf("InstallLockedRegistry read %+v, want the downloaded registry", installed.PackageInfos)
	}
}

// A file whose checksum isn't the recorded one is refused, and nothing is written.
func TestInstaller_DownloadRegistry_checksumMismatch(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	chks := checksum.New()
	other := sha256Of(t, "packages: []\n")
	chks.Set(checksum.RegistryID(rgst), recorded(rgst, other))
	if _, err := newInstaller(rootDir, &servingDownloader{content: servedRegistry}).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, "aqua.yaml", chks); err == nil {
		t.Fatal("an error must be returned")
	}
	if _, err := os.Stat(filepath.Join(rootDir, "registries")); !os.IsNotExist(err) {
		t.Error("a refused file must not be written")
	}
}

// A file already kept under the checksum isn't written again: its content is fixed by the
// checksum, and on Windows a file another aqua is reading can't be replaced.
func TestInstaller_DownloadRegistry_alreadyKept(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	p := lockedPath(t, rootDir, rgst, sha256Of(t, servedRegistry))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	const kept = `{"packages":[{"type":"github_release","repo_owner":"suzuki-shunsuke","repo_name":"kept"}]}`
	if err := os.WriteFile(p+".json", []byte(kept), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newInstaller(rootDir, &servingDownloader{content: servedRegistry}).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, "aqua.yaml", checksum.New()); err != nil {
		t.Fatal(err)
	}
	assertContent(t, p+".json", kept)
	assertContent(t, p, servedRegistry)
}

// InstallLockedRegistry downloads a registry not kept yet, checks it against the recorded
// checksum, and keeps it under that checksum. The next call reads it from disk.
func TestInstaller_InstallLockedRegistry(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	sum := sha256Of(t, servedRegistry)
	d := &servingDownloader{content: servedRegistry}
	inst := newInstaller(rootDir, d)
	for range 2 {
		got, err := inst.InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, recorded(rgst, strings.ToUpper(sum)))
		if err != nil {
			t.Fatal(err)
		}
		if len(got.PackageInfos) != 1 || got.PackageInfos[0].RepoName != "ci-info" {
			t.Errorf("got %+v, want the served registry", got.PackageInfos)
		}
	}
	if d.calls != 1 {
		t.Errorf("downloaded %d times, want once", d.calls)
	}
	assertContent(t, lockedPath(t, rootDir, rgst, sum), servedRegistry)
	assertNoTemporaryFiles(t, rootDir)
}

// What the ref serves now doesn't match the recorded checksum: it is refused and not kept.
func TestInstaller_InstallLockedRegistry_checksumMismatch(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	other := sha256Of(t, "packages: []\n")
	if _, err := newInstaller(rootDir, &servingDownloader{content: servedRegistry}).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, recorded(rgst, other)); err == nil {
		t.Fatal("an error must be returned")
	}
	if _, err := os.Stat(filepath.Join(rootDir, "registries")); !os.IsNotExist(err) {
		t.Error("a refused file must not be written")
	}
}

// A checksum that isn't a hex digest never becomes part of a path.
func TestInstaller_InstallLockedRegistry_invalidChecksum(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	d := &servingDownloader{content: servedRegistry}
	if _, err := newInstaller(rootDir, d).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, recorded(rgst, "../../../../github_content")); err == nil {
		t.Fatal("an error must be returned")
	}
	if d.calls != 0 {
		t.Error("nothing must be downloaded")
	}
}

// writeStale installs a registry file and its JSON that don't match what is served.
func writeStale(t *testing.T, p string) {
	t.Helper()
	for path, content := range map[string]string{p: "edited\n", p + ".json": `{"packages": []}`} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func checkRecorded(t *testing.T, chks *checksum.Checksums, rgst *aqua.Registry, sum string) {
	t.Helper()
	if c := chks.Get(checksum.RegistryID(rgst)); c == nil || !strings.EqualFold(c.Checksum, sum) {
		t.Errorf("recorded %+v, want the checksum of the served file", c)
	}
}

func assertContent(t *testing.T, p, want string) {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != want {
		t.Errorf("%s is %q, want %q", p, b, want)
	}
}

func assertNoTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	if err := filepath.WalkDir(dir, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(p, ".tmp") {
			t.Errorf("a temporary file is left: %s", p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
