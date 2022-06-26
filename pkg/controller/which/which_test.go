package which_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	"github.com/clivm/clivm/pkg/config/aqua"
	cfgRegistry "github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/controller/which"
	"github.com/clivm/clivm/pkg/download"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func stringP(s string) *string {
	return &s
}

func Test_controller_Which(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		files   map[string]string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		isErr   bool
		exp     *which.Which
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
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			exeName: "aqua-installer",
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			exp: &which.Which{
				Package: &config.Package{
					Package: &aqua.Package{
						Name:     "clivm/clivm-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-installer",
						Path:      stringP("aqua-installer"),
					},
				},
				File: &cfgRegistry.File{
					Name: "aqua-installer",
				},
				ExePath: "/home/foo/.local/share/clivm/pkgs/github_content/github.com/clivm/clivm-installer/v1.0.0/aqua-installer/aqua-installer",
			},
		},
		{
			name: "outside aqua",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/clivm/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"/usr/local/foo/gh": "",
			},
			links: map[string]string{
				"../foo/gh": "/usr/local/bin/gh",
			},
			exp: &which.Which{
				ExePath: "/usr/local/bin/gh",
			},
		},
		{
			name: "global config",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:                   "/home/foo/workspace",
				RootDir:               "/home/foo/.local/share/clivm",
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{"/etc/aqua/aqua.yaml"},
			},
			exeName: "aqua-installer",
			files: map[string]string{
				"/etc/aqua/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: suzuki-shunsuke/ci-info@v1.0.0
- name: clivm/clivm-installer@v1.0.0
`,
				"/etc/aqua/registry.yaml": `packages:
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
  supported_if: "false"
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: github-comment
  asset: "github-comment_{{.Arch}}-{{.OS}}.tar.gz"
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			exp: &which.Which{
				Package: &config.Package{
					Package: &aqua.Package{
						Name:     "clivm/clivm-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-installer",
						Path:      stringP("aqua-installer"),
					},
				},
				File: &cfgRegistry.File{
					Name: "aqua-installer",
				},
				ExePath: "/home/foo/.local/share/clivm/pkgs/github_content/github.com/clivm/clivm-installer/v1.0.0/aqua-installer/aqua-installer",
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
			linker := link.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewRegistryDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs), registry.New(d.param, downloader, fs), d.rt, osenv.NewMock(d.env), fs, linker)
			which, err := ctrl.Which(ctx, d.param, d.exeName, logE)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, which); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
