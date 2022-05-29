package generate_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/controller/generate"
	"github.com/aquaproj/aqua/pkg/download"
	githubSvc "github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-github/v44/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_controller_Generate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		files          map[string]string
		links          map[string]string
		args           []string
		env            map[string]string
		param          *config.Param
		rt             *runtime.Runtime
		isErr          bool
		idxs           []int
		fuzzyFinderErr error
		releases       []*github.RepositoryRelease
		tags           []string
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
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			idxs: []int{0},
			releases: []*github.RepositoryRelease{
				{
					TagName: stringP("v1.0.0"),
				},
			},
		},
		{
			name: "arg",
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
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			args: []string{
				"aquaproj/aqua-installer",
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: stringP("v1.0.0"),
				},
			},
		},
		{
			name: "file",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
				File:           "list.txt",
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"list.txt": "aquaproj/aqua-installer\n",
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: stringP("v1.0.0"),
				},
			},
		},
		{
			name: "version filter",
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
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"registry.yaml": `packages:
- type: github_release
  repo_owner: kubernetes-sigs
  repo_name: kustomize
  asset: 'kustomize_{{trimPrefix "kustomize/" .Version}}_{{.OS}}_{{.Arch}}.tar.gz'
  version_filter: 'Version startsWith "kustomize/"'
`,
			},
			args: []string{"kubernetes-sigs/kustomize"},
			releases: []*github.RepositoryRelease{
				{
					TagName: stringP("v4.0.0"),
				},
				{
					TagName: stringP("kustomize/v4.2.0"),
				},
			},
		},
		{
			name: "generate insert",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
				Insert:         true,
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			args: []string{
				"aquaproj/aqua-installer",
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: stringP("v1.0.0"),
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
			configFinder := finder.NewConfigFinder(fs)
			gh := githubSvc.NewMock(d.releases, nil, "")
			v4Client := githubSvc.NewMockGraphQL(d.tags, nil)
			downloader := download.NewRegistryDownloader(gh, download.NewHTTPDownloader(http.DefaultClient))
			registryInstaller := registry.New(d.param, downloader, fs)
			configReader := reader.New(fs)
			fuzzyFinder := generate.NewMockFuzzyFinder(d.idxs, d.fuzzyFinderErr)
			ctrl := generate.New(configFinder, configReader, registryInstaller, gh, fs, fuzzyFinder, v4Client)
			if err := ctrl.Generate(ctx, logE, d.param, d.args...); err != nil {
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
