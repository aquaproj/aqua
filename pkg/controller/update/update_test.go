package update_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/update"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	rgst "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestController_Update(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		isErr          bool
		files          map[string]string
		expFiles       map[string]string
		param          *config.Param
		releases       []*github.RepositoryRelease
		tags           []*github.RepositoryTag
		rt             *runtime.Runtime
		idxs           []int
		fuzzyFinderErr error
		findResults    map[string]*which.FindResult
		registries     map[string]*registry.Config
		versions       map[string]string
	}{
		{
			name: "update commands",
			param: &config.Param{
				Args: []string{"tfcmt", "gh"},
			},
			versions: map[string]string{
				"suzuki-shunsuke/tfcmt": "v4.0.0",
				"cli/cli":               "v2.30.0",
			},
			files: map[string]string{
				"/workspace/aqua.yaml": `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				"/workspace/aqua.yaml": `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.30.0
`,
			},
			findResults: map[string]*which.FindResult{
				"tfcmt": {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     "suzuki-shunsuke/tfcmt",
							Registry: "standard",
							Version:  "v3.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "tfcmt",
							Asset:     "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
						Registry: &aqua.Registry{
							Name:      "standard",
							Type:      "github_content",
							RepoOwner: "aquaproj",
							RepoName:  "aqua-registry",
							Ref:       "v4.0.0",
							Path:      "registry.yaml",
						},
					},
					File: &registry.File{
						Name: "tfcmt",
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/tfcmt/v3.0.0/tfcmt_darwin_arm64.tar.gz/tfcmt",
					ConfigFilePath: "/workspace/aqua.yaml",
				},
				"gh": {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     "cli/cli",
							Registry: "standard",
							Version:  "v2.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      "github_release",
							RepoOwner: "cli",
							RepoName:  "cli",
							Asset:     "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip",
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  "{{.AssetWithoutExt}}/bin/gh",
								},
							},
						},
						Registry: &aqua.Registry{
							Name:      "standard",
							Type:      "github_content",
							RepoOwner: "aquaproj",
							RepoName:  "aqua-registry",
							Ref:       "v4.0.0",
							Path:      "registry.yaml",
						},
					},
					File: &registry.File{
						Name: "gh",
						Src:  "{{.AssetWithoutExt}}/bin/gh",
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/cli/cli/v2.0.0/gh_2.0.0_macOS_arm64.zip/gh_2.0.0_macOS_arm64/bin/gh",
					ConfigFilePath: "/workspace/aqua.yaml",
				},
			},
		},
		{
			name: "no arg",
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			param: &config.Param{
				PWD: "/workspace",
			},
			versions: map[string]string{
				"suzuki-shunsuke/tfcmt": "v4.0.0",
				"cli/cli":               "v2.30.0",
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "tfcmt",
							Asset:     "tfcmt_{{.OS}}_{{.Arch}}.tar.gz",
						},
						{
							Type:      "github_release",
							RepoOwner: "cli",
							RepoName:  "cli",
							Asset:     "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip",
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  "{{.AssetWithoutExt}}/bin/gh",
								},
							},
						},
					},
				},
			},
			files: map[string]string{
				"/workspace/aqua.yaml": `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				"/workspace/aqua.yaml": `registries:
- type: standard
  ref: v4.60.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.30.0
`,
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: ptr.String("v4.60.0"),
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
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
				Tags:     d.tags,
			}
			configReader := reader.New(fs, d.param)
			configFinder := finder.NewConfigFinder(fs)
			registryInstaller := &rgst.MockInstaller{
				M: d.registries,
			}
			fuzzyFinder := fuzzyfinder.NewMock(d.idxs, d.fuzzyFinderErr)
			whichCtrl := &which.MockMultiController{
				FindResults: d.findResults,
			}
			fuzzyGetter := versiongetter.NewMockFuzzyGetter(d.versions)
			ctrl := update.New(d.param, gh, configFinder, configReader, registryInstaller, fs, d.rt, fuzzyGetter, fuzzyFinder, whichCtrl)
			if err := ctrl.Update(ctx, logE, d.param); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			for path, expBody := range d.expFiles {
				b, err := afero.ReadFile(fs, path)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(expBody, string(b)); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
