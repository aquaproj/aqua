package installpackage_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func Test_installer_InstallAqua(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name               string
		files              map[string]string
		param              *config.Param
		rt                 *runtime.Runtime
		checksumDownloader download.ChecksumDownloader
		checksumCalculator installpackage.ChecksumCalculator
		version            string
		isTest             bool
		isErr              bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			files: map[string]string{
				"/home/foo/.local/share/aquaproj-aqua/internal/pkgs/github_release/github.com/aquaproj/aqua/v1.6.1/aqua_linux_amd64.tar.gz/aqua": "xxx",
			},
			version: "v1.6.1",
			checksumCalculator: &installpackage.MockChecksumCalculator{
				Checksum: "c6f3b1f37d9bf4f73e6c6dcf1bd4bb59b48447ad46d4b72e587d15f66a96ab5a",
			},
			checksumDownloader: &download.MockChecksumDownloader{
				Body: `31adc2cfc3aab8e66803f6769016fe6953a22f88de403211abac83c04a542d46  aqua_darwin_arm64.tar.gz
6e53f151abf10730bdfd4a52b99019ffa5f58d8ad076802affb3935dd82aba96  aqua_darwin_amd64.tar.gz
c6f3b1f37d9bf4f73e6c6dcf1bd4bb59b48447ad46d4b72e587d15f66a96ab5a  aqua_linux_amd64.tar.gz
e922723678f493216c2398f3f23fb027c9a98808b49f6fce401ef82ee2c22b03  aqua_linux_arm64.tar.gz`,
				Code: 200,
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			ctrl := installpackage.New(d.param, &download.Mock{
				RC: io.NopCloser(strings.NewReader("xxx")),
			}, d.rt, fs, domain.NewMockLinker(fs), d.checksumDownloader, d.checksumCalculator, &unarchive.MockUnarchiver{}, &policy.Checker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockPipInstaller{}, &installpackage.MockCargoPackageInstaller{})
			if err := ctrl.InstallAqua(ctx, logE, d.version); err != nil {
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
