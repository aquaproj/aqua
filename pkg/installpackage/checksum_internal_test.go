package installpackage

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestInstallerImpl_verifyChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		param *ParamVerifyChecksum
		isErr bool
		inst  *InstallerImpl
	}{
		{
			name: "normal",
			param: &ParamVerifyChecksum{
				AssetName: "gh_2.17.0_macOS_amd64.tar.gz",
				Pkg: &config.Package{
					PackageInfo: &registry.PackageInfo{
						Type:     "github_release",
						Rosetta2: ptr.Bool(true),
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
				Checksums:    checksum.New(),
				ChecksumID:   "github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_darwin_arm64.tar.gz",
				TempFilePath: "/tmp/verify_checksum/tempfile",
			},
			inst: &InstallerImpl{
				fs: afero.NewMemMapFs(),
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
				runtime: &runtime.Runtime{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				checksumCalculator: &MockChecksumCalculator{
					Checksum: "3516a4d84f7b69ea5752ca2416895a2705910af3ed6815502af789000fc7e963",
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

			if err := inst.verifyChecksum(ctx, logE, d.param); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
		})
	}
}
