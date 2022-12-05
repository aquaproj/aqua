package installpackage_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func Test_installer_InstallProxy(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		rt       *runtime.Runtime
		executor installpackage.Executor
		links    map[string]string
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				fmt.Sprintf("/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion): "",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			linker := domain.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, d.executor, nil, &checksum.Calculator{}, unarchive.New(), &domain.MockPolicyChecker{}, &domain.MockCosignVerifier{})
			if err := ctrl.InstallProxy(ctx, logE); err != nil {
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
