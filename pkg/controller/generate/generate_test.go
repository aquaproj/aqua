package generate_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/controller/generate"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_controller_Generate(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()
	data := []struct {
		name               string
		files              map[string]string
		links              map[string]string
		args               []string
		env                map[string]string
		param              *config.Param
		rt                 *runtime.Runtime
		isErr              bool
		idxs               []int
		idx                int
		fuzzyFinderErr     error
		versionSelectorErr error
		releases           []*github.RepositoryRelease
		tags               []*github.RepositoryTag
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
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
			name: "select-version (release)",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
				SelectVersion:  true,
			},
			idx: 1,
			files: map[string]string{
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
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
					TagName: stringP("v1.1.0"),
				},
				{
					TagName: stringP("v1.0.0"),
				},
			},
		},
		{
			name: "select-version (tag)",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
				SelectVersion:  true,
			},
			idx: 1,
			files: map[string]string{
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
  version_source: github_tag
`,
			},
			args: []string{
				"aquaproj/aqua-installer",
			},
			tags: []*github.RepositoryTag{
				{
					Name: stringP("v1.1.0"),
				},
				{
					Name: stringP("v1.0.0"),
				},
			},
		},
		{
			name: "github_tag",
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
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
  version_source: github_tag
`,
			},
			args: []string{
				"aquaproj/aqua-installer",
			},
			tags: []*github.RepositoryTag{
				{
					Name: stringP("v1.1.0"),
				},
				{
					Name: stringP("go1.0.0"),
				},
			},
		},
		{
			name: "github_tag with version_filter",
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
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
  version_source: github_tag
  version_filter: 'Version startsWith "go"'
`,
			},
			args: []string{
				"aquaproj/aqua-installer",
			},
			tags: []*github.RepositoryTag{
				{
					Name: stringP("v1.1.0"),
				},
				{
					Name: stringP("go1.0.0"),
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
			linker := domain.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			configFinder := finder.NewConfigFinder(fs)
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
				Tags:     d.tags,
			}
			downloader := download.NewGitHubContentFileDownloader(gh, download.NewHTTPDownloader(http.DefaultClient))
			registryInstaller := registry.New(d.param, downloader, fs, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			configReader := reader.New(fs, d.param)
			fuzzyFinder := generate.NewMockFuzzyFinder(d.idxs, d.fuzzyFinderErr)
			versionSelector := generate.NewMockVersionSelector(d.idx, d.versionSelectorErr)
			ctrl := generate.New(configFinder, configReader, registryInstaller, gh, fs, fuzzyFinder, versionSelector)
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
