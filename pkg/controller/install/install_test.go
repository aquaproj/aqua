package install_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/install"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
)

func TestController_Install(t *testing.T) { //nolint:funlen
	t.Parallel()
	// The paths of files, dirs, and links are relative to a home directory
	// created for each test case.
	const (
		workspace = "workspace"
		rootDir   = ".local/share/aquaproj-aqua"
	)
	// The path of aqua-proxy relative to the root directory.
	proxyPath := fmt.Sprintf("internal/pkgs/github_release/github.com/aquaproj/aqua-proxy/%s/aqua-proxy_linux_amd64.tar.gz/aqua-proxy", installpackage.ProxyVersion)
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
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				workspace + "/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				workspace + "/aqua-policy.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- registry: standard
`,
				workspace + "/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				rootDir + "/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer": ``,
				rootDir + "/" + proxyPath: ``,
				rootDir + "/aqua-proxy":   ``,
			},
			dirs: []string{
				workspace + "/.git",
				rootDir + "/bin",
			},
			links: map[string]string{
				rootDir + "/bin/aqua-installer": "../aqua-proxy",
				rootDir + "/bin/aqua-proxy":     "../" + proxyPath,
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	registryDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			home := t.TempDir()
			testutil.WriteFiles(t, home, d.files, d.dirs...)
			linker := link.New()
			for src, dest := range d.links {
				if err := linker.Symlink(dest, filepath.Join(home, filepath.FromSlash(src))); err != nil {
					t.Fatal(err)
				}
			}
			d.param.CWD = filepath.Join(home, workspace)
			d.param.RootDir = filepath.Join(home, filepath.FromSlash(rootDir))

			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			executor := &osexec.Mock{}
			vacuumMock := vacuum.NewMock(d.param.RootDir, nil, nil)
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, linker, nil, &checksum.Calculator{}, unarchive.New(executor), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &minisign.MockVerifier{}, &ghattestation.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{}, vacuumMock)
			policyFinder := policy.NewConfigFinder()
			policyReader := policy.NewReader(&policy.MockValidator{}, policyFinder, policy.NewConfigReader())
			ctrl := install.New(d.param, finder.NewConfigFinder(), reader.New(d.param), registry.New(d.param, registryDownloader, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), pkgInstaller, d.rt, policyReader)
			if err := ctrl.Install(ctx, logger, d.param); err != nil {
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
