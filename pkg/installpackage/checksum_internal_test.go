package installpackage

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestInstaller_extractChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		pkg          *config.Package
		runtime      *runtime.Runtime
		assetName    string
		checksumFile []byte
		isErr        bool
		c            string
	}{
		{
			name: "raw",
			runtime: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
			assetName:    "spotify-tui-windows.sha256",
			checksumFile: []byte("423b7c1a842bb4cd847b492cad2d1c724e00f92ed913226e9ac1ab0925b0b639"),
			c:            "423b7c1a842bb4cd847b492cad2d1c724e00f92ed913226e9ac1ab0925b0b639",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "Rigellute/spotify-tui",
					Version: "v0.25.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "Rigellute",
					RepoName:  "spotify-tui",
					Asset:     strP("spotify-tui-{{.OS}}.tar.gz"),
					Checksum: &registry.Checksum{
						Type:       "github_release",
						FileFormat: "raw",
						Asset:      "spotify-tui-{{.OS}}.sha256",
						Algorithm:  "sha256",
					},
				},
			},
		},
		{
			name: "colima",
			runtime: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			assetName:    "colima-Darwin-arm64",
			checksumFile: []byte("8d5b5f66ec04a4e21d147e2b7501d44b084eacf3dc03af081e681fb99697c3ac  colima-Darwin-arm64\n"),
			c:            "8d5b5f66ec04a4e21d147e2b7501d44b084eacf3dc03af081e681fb99697c3ac",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "abiosoft/colima",
					Version: "v0.4.4",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "abiosoft",
					RepoName:  "colima",
					Asset:     strP("colima-{{.OS}}-{{.Arch}}"),
					Replacements: map[string]string{
						"darwin": "Darwin",
						"linux":  "Linux",
						"amd64":  "x86_64",
					},
					Checksum: &registry.Checksum{
						Type:       "github_release",
						FileFormat: "regexp",
						Asset:      "colima-{{.OS}}-{{.Arch}}.sha256sum",
						Algorithm:  "sha256",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(.{64})`,
							File:     `^.{64}\s+(\S+)$`,
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			inst := &Installer{
				runtime: &runtime.Runtime{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
			}
			if d.runtime != nil {
				inst.runtime = d.runtime
			}
			c, err := inst.extractChecksum(d.pkg, d.assetName, d.checksumFile)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if c != d.c {
				t.Fatalf("checksum wanted %s, got %s", d.c, c)
			}
		})
	}
}

type mockChecksumDownloader struct {
	body string
	code int64
	err  error
}

func (dl *mockChecksumDownloader) DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error) {
	if dl.err != nil {
		return nil, dl.code, dl.err
	}
	return io.NopCloser(strings.NewReader(dl.body)), dl.code, dl.err
}

type mockChecksumCalculator struct {
	checksum string
	err      error
}

func (calc *mockChecksumCalculator) Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	return calc.checksum, calc.err
}

func boolP(b bool) *bool {
	return &b
}

func TestInstaller_verifyChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		param *ParamVerifyChecksum
		isErr bool
		inst  *Installer
	}{
		{
			name: "normal",
			param: &ParamVerifyChecksum{
				AssetName: "gh_2.17.0_macOS_amd64.tar.gz",
				Pkg: &config.Package{
					PackageInfo: &registry.PackageInfo{
						Type:     "github_release",
						Rosetta2: boolP(true),
						Checksum: &registry.Checksum{
							Type:       "github_release",
							Algorithm:  "sha256",
							FileFormat: "regexp",
							Pattern: &registry.ChecksumPattern{
								Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
								File:     "^\\b[A-Fa-f0-9]{64}\\b\\s+(\\S+)$",
							},
						},
						Replacements: map[string]string{
							"darwin": "macOS",
						},
					},
				},
				Checksums:  checksum.New(),
				ChecksumID: "github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_darwin_arm64.tar.gz",
				TempDir:    "/tmp/verify_checksum",
				Body:       io.NopCloser(strings.NewReader("")),
			},
			inst: &Installer{
				fs: afero.NewMemMapFs(),
				checksumDownloader: &mockChecksumDownloader{
					body: `2005b4aef5fec0336cb552c74f3e4c445dcdd9e9c1e217d8de3acd45ee152470  gh_2.17.0_linux_386.deb
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
				runtime: &runtime.Runtime{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				checksumFileParser: &checksum.FileParser{},
				checksumCalculator: &mockChecksumCalculator{
					checksum: "3516a4d84f7b69ea5752ca2416895a2705910af3ed6815502af789000fc7e963",
				},
			},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			inst := d.inst
			rc, err := inst.verifyChecksum(ctx, logE, d.param)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			defer rc.Close()
			if d.isErr {
				t.Fatal("error must occur")
			}
		})
	}
}
