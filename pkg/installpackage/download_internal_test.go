package installpackage

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/spf13/afero"
)

// writeUnarchiver stands in for a real unarchiver by writing exeName into the
// destination it is handed, so that tests can tell which directory the
// extraction went to.
type writeUnarchiver struct {
	fs      afero.Fs
	exeName string
	err     error
}

func (u *writeUnarchiver) Unarchive(_ context.Context, _ *slog.Logger, _ *unarchive.File, dest string) error {
	if u.err != nil {
		return u.err
	}
	p := filepath.Join(dest, u.exeName)
	if err := u.fs.MkdirAll(filepath.Dir(p), dirPerm); err != nil {
		return err //nolint:wrapcheck
	}
	return afero.WriteFile(u.fs, p, []byte("gh"), dirPerm) //nolint:wrapcheck
}

const dirPerm = 0o755

// newUnarchiveTestInstaller returns an installer rooted at a fresh temporary
// directory, along with that root and the package destination under it.
func newUnarchiveTestInstaller(t *testing.T, unarchiveErr error) (*Installer, string, string) {
	t.Helper()
	fs := afero.NewOsFs()
	rootDir := t.TempDir()
	dest := filepath.Join(rootDir, "pkgs", "github_release", "github.com", "cli", "cli", "v2.96.0", "gh_2.96.0_linux_amd64.tar.gz")
	return &Installer{
		fs:      fs,
		rootDir: rootDir,
		unarchiver: &writeUnarchiver{
			fs:      fs,
			exeName: filepath.Join("bin", "gh"),
			err:     unarchiveErr,
		},
	}, rootDir, dest
}

// assertTempDirIsEmpty checks that neither a discarded nor a failed extraction
// left a temporary directory behind.
func assertTempDirIsEmpty(t *testing.T, fs afero.Fs, rootDir string) {
	t.Helper()
	entries, err := afero.ReadDir(fs, filepath.Join(rootDir, "temp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("%d temporary directories are left", len(entries))
	}
}

// The OS filesystem is used rather than a memory-mapped one because these tests
// turn on Rename's real behaviour: MemMapFs happily renames onto an existing
// directory, which is exactly the case the lost-race test needs to exercise.
func TestInstaller_unarchive(t *testing.T) {
	t.Parallel()

	inst, rootDir, dest := newUnarchiveTestInstaller(t, nil)

	if err := inst.unarchive(t.Context(), slog.New(slog.DiscardHandler), &DownloadParam{
		Dest:  dest,
		Asset: "gh_2.96.0_linux_amd64.tar.gz",
	}, nil, "tar.gz"); err != nil {
		t.Fatal(err)
	}

	b, err := afero.ReadFile(inst.fs, filepath.Join(dest, "bin", "gh"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "gh" {
		t.Fatalf("the executable file is %q, want %q", b, "gh")
	}
	assertTempDirIsEmpty(t, inst.fs, rootDir)
}

// A package another process finished extracting first must be left alone, and
// this installer's own copy discarded.
func TestInstaller_unarchive_destExists(t *testing.T) {
	t.Parallel()

	inst, rootDir, dest := newUnarchiveTestInstaller(t, nil)
	exePath := filepath.Join(dest, "bin", "gh")
	if err := inst.fs.MkdirAll(filepath.Dir(exePath), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(inst.fs, exePath, []byte("installed by someone else"), dirPerm); err != nil {
		t.Fatal(err)
	}

	if err := inst.unarchive(t.Context(), slog.New(slog.DiscardHandler), &DownloadParam{
		Dest:  dest,
		Asset: "gh_2.96.0_linux_amd64.tar.gz",
	}, nil, "tar.gz"); err != nil {
		t.Fatal(err)
	}

	b, err := afero.ReadFile(inst.fs, exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "installed by someone else" {
		t.Fatalf("the executable file is %q, want %q", b, "installed by someone else")
	}
	assertTempDirIsEmpty(t, inst.fs, rootDir)
}

// The destination must not be created at all when the extraction fails,
// otherwise downloadWithRetry would treat the package as installed.
func TestInstaller_unarchive_failure(t *testing.T) {
	t.Parallel()

	inst, rootDir, dest := newUnarchiveTestInstaller(t, errors.New("unarchive failed"))

	if err := inst.unarchive(t.Context(), slog.New(slog.DiscardHandler), &DownloadParam{
		Dest:  dest,
		Asset: "gh_2.96.0_linux_amd64.tar.gz",
	}, nil, "tar.gz"); err == nil {
		t.Fatal("error must be returned")
	}

	if f, err := afero.Exists(inst.fs, dest); err != nil {
		t.Fatal(err)
	} else if f {
		t.Fatal("the destination must not be created when the extraction fails")
	}
	assertTempDirIsEmpty(t, inst.fs, rootDir)
}

func TestInstaller_download(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		param *DownloadParam
		inst  *Installer
		isErr bool
	}{
		{
			name: "normal",
			param: &DownloadParam{
				Package: &config.Package{
					Package: &aqua.Package{
						Name:    "cli/cli",
						Version: "v2.17.0",
					},
					PackageInfo: &registry.PackageInfo{
						Type:      pkgTypeGitHubRelease,
						RepoOwner: "cli",
						RepoName:  "cli",
						Asset:     "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
						Rosetta2:  true,
						Checksum: &registry.Checksum{
							Type:       pkgTypeGitHubRelease,
							Algorithm:  algoSHA256,
							FileFormat: algoTypeRegexp,
							Pattern: &registry.ChecksumPattern{
								Checksum: regexpSHA256,
								File:     "^\\b[A-Fa-f0-9]{64}\\b\\s+(\\S+)$",
							},
						},
					},
				},
				Checksums: checksum.New(),
				Asset:     "gh_2.17.0_macOS_amd64.tar.gz",
			},
			inst: &Installer{
				progressBar: true,
				runtime: &runtime.Runtime{
					GOOS:   osDarwin,
					GOARCH: "arm64",
				},
				fs: afero.NewMemMapFs(),
				downloader: &download.Mock{
					RC: io.NopCloser(strings.NewReader("hello")),
				},
				unarchiver: &unarchive.MockUnarchiver{},
				checksumCalculator: &MockChecksumCalculator{
					Checksum: "3516a4d84f7b69ea5752ca2416895a2705910af3ed6815502af789000fc7e963",
				},
				checksumDownloader: &download.MockChecksumDownloader{
					Body: `2005b4aef5fec0336cb552c74f3e4c445dcdd9e9c1e217d8de3acd45ee152470  gh_2.17.0_linux_386.deb
34c0ba49d290ffe108c723ffb0063a4a749a8810979b71fc503434b839688b5c  gh_2.17.0_linux_386.rpm
3516a4d84f7b69ea5752ca2416895a2705910af3ed6815502af789000fc7e963  gh_2.17.0_macOS_amd64.tar.gz
3fb9532fd907547ad1ed89d507f785589c70f3896133ca64de609ba0dcc080d5  gh_2.17.0_linux_armv6.tar.gz
4bd7415b5ccc559b2e9ff7d4bcb8d1fd63c4acce3eaf589da2a70c50035af54f  gh_2.17.0_linux_amd64.deb
5859178d22f0124bbedc8d69c242df8c304ba8da1eb94406f11b1bbe4ec393e8  gh_2.17.0_linux_amd64.rpm
8c403207ed8ab18b4c69d7e97321a553731d9034fe98ba96feebfc267ecd2c91  gh_2.17.0_linux_armv6.deb
96d4e523636446b796b28f069332b6f8ea9a0950c6ef43617203cc5ac5af0d84  gh_2.17.0_windows_amd64.zip
a614f898e229f3d6af3cea88cb42ff71c4c5466a52fefef2118d307f1a11b055  gh_2.17.0_linux_armv6.rpm
c36f5ead31b8d6c41dc5ce97b514133a8cc037739aba239aa2a75b8afe3e618a  gh_2.17.0_linux_arm64.deb
c6ce28981a1fb9acb13ee091b5f3de8eb244a67dc99aff1d106985c1e94c72c6  gh_2.17.0_linux_amd64.tar.gz
cdd97a4afe4ec828fed72811f9b47a9fa4ef8f8fb2fa1e3b9a8cfc3334cbc815  gh_2.17.0_linux_arm64.rpm
d373e305512e53145df7064a0253df696fe17f9ec71804311239f3e2c9e19999  gh_2.17.0_linux_arm64.tar.gz
d3b06f291551ce0357e08334d8ba72810a552b593329e3c0dd3489f51a8712a3  gh_2.17.0_windows_386.zip
ed2ed654e1afb92e5292a43213e17ecb0fe0ec50c19fe69f0d185316a17d39fa  gh_2.17.0_linux_386.tar.gz`,
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			if err := d.inst.download(ctx, logger, d.param); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
