package install_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/controller/install"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/exec"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/aquaproj/aqua/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestController_Install(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		links             map[string]string
		param             *config.Param
		rt                *runtime.Runtime
		registryInstaller registry.Installer
		isErr             bool
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer":                                              ``,
				fmt.Sprintf("/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion): ``,
				"/home/foo/.local/share/aquaproj-aqua/bin/aqua-installer": ``,
				"/home/foo/.local/share/aquaproj-aqua/bin/aqua-proxy":     ``,
			},
			links: map[string]string{
				"aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/aqua-installer",
				fmt.Sprintf("../pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion): "/home/foo/.local/share/aquaproj-aqua/bin/aqua-proxy",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	registryDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
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
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			executor := &exec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, executor, nil, &checksum.Calculator{}, unarchive.New(), &domain.MockPolicyChecker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			ctrl := install.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, registryDownloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), pkgInstaller, fs, d.rt, &policy.MockConfigReader{}, &domain.MockCosignInstaller{})
			if err := ctrl.Install(ctx, logE, d.param); err != nil {
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
