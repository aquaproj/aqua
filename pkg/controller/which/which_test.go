package which_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	cfgRegistry "github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/aquaproj/aqua/pkg/download"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
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
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			exeName: "aqua-installer",
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
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
						Name:     "aquaproj/aqua-installer",
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
				ExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer",
				Config: &aqua.Config{
					Registries: aqua.Registries{
						"standard": {
							Name: "standard",
							Type: "local",
							Path: "registry.yaml",
						},
					},
					Packages: []*aqua.Package{
						{
							Name:     "aquaproj/aqua-installer",
							Registry: "standard",
							Version:  "v1.0.0",
						},
					},
				},
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
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/aquaproj-aqua/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
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
				RootDir:               "/home/foo/.local/share/aquaproj-aqua",
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
- name: aquaproj/aqua-installer@v1.0.0
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
						Name:     "aquaproj/aqua-installer",
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
				ExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer",
				Config: &aqua.Config{
					Registries: aqua.Registries{
						"standard": {
							Name: "standard",
							Type: "local",
							Path: "registry.yaml",
						},
					},
					Packages: []*aqua.Package{
						{
							Name:     "suzuki-shunsuke/ci-info",
							Registry: "standard",
							Version:  "v1.0.0",
						},
						{
							Name:     "aquaproj/aqua-installer",
							Registry: "standard",
							Version:  "v1.0.0",
						},
					},
				},
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
			downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
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
