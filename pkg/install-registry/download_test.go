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

var commit1 = strings.Repeat("1", 40) //nolint:gochecknoglobals

// servingDownloader serves the registry file at the commits it knows, and remembers the
// refs it was asked for. With offline, it fails instead, to show a registry was read from
// disk.
type servingDownloader struct {
	contents map[string]string
	offline  bool
	refs     []string
}

// serving serves servedRegistry at commit1.
func serving() *servingDownloader {
	return &servingDownloader{contents: map[string]string{commit1: servedRegistry}}
}

func (d *servingDownloader) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	d.refs = append(d.refs, param.Ref)
	if d.offline {
		return nil, errors.New("offline")
	}
	content, ok := d.contents[param.Ref]
	if !ok {
		return nil, errors.New("not found")
	}
	return &domain.GitHubContentFile{String: content}, nil
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

func commitPath(t *testing.T, rootDir string, rgst *aqua.Registry, commit string) string {
	t.Helper()
	p, err := rgst.CommitFilePath(rootDir, commit)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func recorded(rgst *aqua.Registry, sum string) *checksum.Checksum {
	id := checksum.RegistryID(rgst)
	return &checksum.Checksum{ID: id, Checksum: sum, Algorithm: "sha256"}
}

// DownloadRegistry downloads the registry from the commit even when a file is installed at
// the ref's path, records its checksum, and keeps it under the commit, with the JSON made
// from it. The file at the ref's path is left alone.
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

	d := serving()
	chks := checksum.New()
	got, err := newInstaller(rootDir, d).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, commit1, chks)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PackageInfos) != 1 || got.PackageInfos[0].RepoName != "ci-info" {
		t.Errorf("got %+v, want the served registry", got.PackageInfos)
	}
	if len(d.refs) != 1 || d.refs[0] != commit1 {
		t.Errorf("downloaded at %v, want the commit", d.refs)
	}
	checkRecorded(t, chks, rgst, sha256Of(t, servedRegistry))
	p := commitPath(t, rootDir, rgst, commit1)
	assertContent(t, p, servedRegistry)
	if _, err := os.Stat(p + ".json"); err != nil {
		t.Errorf("the JSON made from the registry must be kept beside it: %v", err)
	}
	assertContent(t, old, "edited\n")
	assertNoTemporaryFiles(t, rootDir)
	assertRegistryFilesArePrivate(t, filepath.Join(rootDir, "registries"))
	assertReadsOffline(t, rootDir, rgst, sha256Of(t, servedRegistry))
}

// assertReadsOffline checks InstallLockedRegistry reads the kept registry without
// downloading it again.
func assertReadsOffline(t *testing.T, rootDir string, rgst *aqua.Registry, sum string) {
	t.Helper()
	offline := &servingDownloader{offline: true}
	installed, err := newInstaller(rootDir, offline).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, commit1, recorded(rgst, sum))
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
	chks.Set(checksum.RegistryID(rgst), recorded(rgst, sha256Of(t, "packages: []\n")))
	if _, err := newInstaller(rootDir, serving()).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, commit1, chks); err == nil {
		t.Fatal("an error must be returned")
	}
	if _, err := os.Stat(filepath.Join(rootDir, "registries")); !os.IsNotExist(err) {
		t.Error("a refused file must not be written")
	}
}

// A file already kept under the commit isn't written again: its content is fixed by the
// commit, and on Windows a file another aqua is reading can't be replaced.
func TestInstaller_DownloadRegistry_alreadyKept(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	p := commitPath(t, rootDir, rgst, commit1)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	const kept = `{"packages":[{"type":"github_release","repo_owner":"suzuki-shunsuke","repo_name":"kept"}]}`
	if err := os.WriteFile(p+".json", []byte(kept), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newInstaller(rootDir, serving()).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, commit1, checksum.New()); err != nil {
		t.Fatal(err)
	}
	assertContent(t, p+".json", kept)
	assertContent(t, p, servedRegistry)
}

// InstallLockedRegistry downloads a registry not kept yet from the recorded commit, checks
// it against the recorded checksum, and keeps it under the commit. The next call reads it
// from disk.
func TestInstaller_InstallLockedRegistry(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	sum := sha256Of(t, servedRegistry)
	d := serving()
	inst := newInstaller(rootDir, d)
	for range 2 {
		got, err := inst.InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, strings.ToUpper(commit1), recorded(rgst, strings.ToUpper(sum)))
		if err != nil {
			t.Fatal(err)
		}
		if len(got.PackageInfos) != 1 || got.PackageInfos[0].RepoName != "ci-info" {
			t.Errorf("got %+v, want the served registry", got.PackageInfos)
		}
	}
	if len(d.refs) != 1 || !strings.EqualFold(d.refs[0], commit1) {
		t.Errorf("downloaded at %v, want once at the commit", d.refs)
	}
	assertContent(t, commitPath(t, rootDir, rgst, commit1), servedRegistry)
	assertNoTemporaryFiles(t, rootDir)
}

// What the commit holds doesn't match the recorded checksum: it is refused and not kept.
func TestInstaller_InstallLockedRegistry_checksumMismatch(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	if _, err := newInstaller(rootDir, serving()).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, commit1, recorded(rgst, sha256Of(t, "packages: []\n"))); err == nil {
		t.Fatal("an error must be returned")
	}
	if _, err := os.Stat(filepath.Join(rootDir, "registries")); !os.IsNotExist(err) {
		t.Error("a refused file must not be written")
	}
}

// A commit that isn't a commit SHA never becomes part of a path, nor is downloaded.
func TestInstaller_InstallLockedRegistry_invalidCommit(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	rgst := corpRegistry()
	d := serving()
	if _, err := newInstaller(rootDir, d).InstallLockedRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, "../../../../github_content", recorded(rgst, sha256Of(t, servedRegistry))); err == nil {
		t.Fatal("an error must be returned")
	}
	if len(d.refs) != 0 {
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
