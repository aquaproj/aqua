package install_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/install"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/exec"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
)

func TestController_Install(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		dirs              []string
		links             map[string]string
		param             *config.Param
		rt                *runtime.Runtime
		registryInstaller install.RegistryInstaller
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
				"/home/foo/workspace/aqua-policy.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- registry: standard
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer":                                                       ``,
				fmt.Sprintf("/home/foo/.local/share/aquaproj-aqua/internal/pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion): ``,
				"/home/foo/.local/share/aquaproj-aqua/bin/aqua-installer": ``,
				"/home/foo/.local/share/aquaproj-aqua/aqua-proxy":         ``,
			},
			dirs: []string{
				"/home/foo/workspace/.git",
			},
			links: map[string]string{
				"../aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/aqua-installer",
				fmt.Sprintf("../internal/pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion): "/home/foo/.local/share/aquaproj-aqua/bin/aqua-proxy",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	registryDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files, d.dirs...)
			if err != nil {
				t.Fatal(err)
			}
			linker := installpackage.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			executor := &exec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(executor, fs), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{})
			policyFinder := policy.NewConfigFinder(fs)
			policyReader := policy.NewReader(fs, &policy.MockValidator{}, policyFinder, policy.NewConfigReader(fs))
			ctrl := install.New(d.param, finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, registryDownloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), pkgInstaller, fs, d.rt, policyReader)
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
