package update_test

import (
	"log/slog"
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
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestController_Update(t *testing.T) { //nolint:funlen,maintidx
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
		{ //nolint:dupl
			name: "update commands",
			param: &config.Param{
				Args: []string{pkgNameTfcmt, "gh"},
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			files: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.30.0
`,
			},
			findResults: map[string]*which.FindResult{
				pkgNameTfcmt: {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     repoSuzukiTfcmt,
							Registry: regTypeStandard,
							Version:  "v3.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						Registry: &aqua.Registry{
							Name:      regTypeStandard,
							Type:      pkgTypeGitHubContent,
							RepoOwner: repoOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Ref:       "v4.0.0",
							Path:      regFileRegistryYaml,
						},
					},
					File: &registry.File{
						Name: pkgNameTfcmt,
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/tfcmt/v3.0.0/tfcmt_darwin_arm64.tar.gz/tfcmt",
					ConfigFilePath: pathWorkspaceYaml,
				},
				"gh": {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     repoCliCli,
							Registry: regTypeStandard,
							Version:  "v2.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
						Registry: &aqua.Registry{
							Name:      regTypeStandard,
							Type:      pkgTypeGitHubContent,
							RepoOwner: repoOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Ref:       "v4.0.0",
							Path:      regFileRegistryYaml,
						},
					},
					File: &registry.File{
						Name: "gh",
						Src:  tmplGhBinSrc,
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/cli/cli/v2.0.0/gh_2.0.0_macOS_arm64.zip/gh_2.0.0_macOS_arm64/bin/gh",
					ConfigFilePath: pathWorkspaceYaml,
				},
			},
		},
		{
			name: "no arg",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			param: &config.Param{
				CWD: pathWorkspace,
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
					},
				},
			},
			files: map[string]string{
				pathWorkspaceYaml: `checksum:
  enabled: true
registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `checksum:
  enabled: true
registries:
- type: standard
  ref: v4.60.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.30.0
`,
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: new("v4.60.0"),
				},
			},
		},
		{
			name: "only registry",
			param: &config.Param{
				CWD:          pathWorkspace,
				OnlyRegistry: true,
			},
			files: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.60.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: new("v4.60.0"),
				},
			},
		},
		{
			name: "only package",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			param: &config.Param{
				CWD:         pathWorkspace,
				OnlyPackage: true,
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
					},
				},
			},
			files: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.30.0
`,
			},
		},
		{
			name: "select packages",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			param: &config.Param{
				CWD:    pathWorkspace,
				Insert: true,
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
					},
				},
			},
			idxs: []int{1}, // cli/cli
			files: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.30.0
`,
			},
		},
		{
			name: "ignore commit hash",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			param: &config.Param{
				CWD: pathWorkspace,
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
					},
				},
			},
			files: map[string]string{
				pathWorkspaceYaml: `checksum:
  enabled: true
registries:
- type: standard
  ref: 4da26b32f72963f42a04b099d03604dab32c6844
packages:
- name: suzuki-shunsuke/tfcmt@4da26b32f72963f42a04b099d03604dab32c6844
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `checksum:
  enabled: true
registries:
- type: standard
  ref: 4da26b32f72963f42a04b099d03604dab32c6844
packages:
- name: suzuki-shunsuke/tfcmt@4da26b32f72963f42a04b099d03604dab32c6844
- name: cli/cli@v2.30.0
`,
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: new("v4.60.0"),
				},
			},
		},
		{ //nolint:dupl
			name: "update commands with versions",
			param: &config.Param{
				Args: []string{pkgNameTfcmt, "gh@v2.27.0"},
			},
			versions: map[string]string{
				repoSuzukiTfcmt: "v4.0.0",
				repoCliCli:      "v2.30.0",
			},
			files: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v3.0.0
- name: cli/cli@v2.0.0
`,
			},
			expFiles: map[string]string{
				pathWorkspaceYaml: `registries:
- type: standard
  ref: v4.0.0
packages:
- name: suzuki-shunsuke/tfcmt@v4.0.0
- name: cli/cli@v2.27.0
`,
			},
			findResults: map[string]*which.FindResult{
				pkgNameTfcmt: {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     repoSuzukiTfcmt,
							Registry: regTypeStandard,
							Version:  "v3.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerSuzuki,
							RepoName:  pkgNameTfcmt,
							Asset:     tmplTfcmtAsset,
						},
						Registry: &aqua.Registry{
							Name:      regTypeStandard,
							Type:      pkgTypeGitHubContent,
							RepoOwner: repoOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Ref:       "v4.0.0",
							Path:      regFileRegistryYaml,
						},
					},
					File: &registry.File{
						Name: pkgNameTfcmt,
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/tfcmt/v3.0.0/tfcmt_darwin_arm64.tar.gz/tfcmt",
					ConfigFilePath: pathWorkspaceYaml,
				},
				"gh": {
					Package: &config.Package{
						Package: &aqua.Package{
							Name:     repoCliCli,
							Registry: regTypeStandard,
							Version:  "v2.0.0",
						},
						PackageInfo: &registry.PackageInfo{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoOwnerCli,
							RepoName:  repoOwnerCli,
							Asset:     tmplGhAsset,
							Files: []*registry.File{
								{
									Name: "gh",
									Src:  tmplGhBinSrc,
								},
							},
						},
						Registry: &aqua.Registry{
							Name:      regTypeStandard,
							Type:      pkgTypeGitHubContent,
							RepoOwner: repoOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Ref:       "v4.0.0",
							Path:      regFileRegistryYaml,
						},
					},
					File: &registry.File{
						Name: "gh",
						Src:  tmplGhBinSrc,
					},
					Config:         &aqua.Config{},
					ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/cli/cli/v2.0.0/gh_2.0.0_macOS_arm64.zip/gh_2.0.0_macOS_arm64/bin/gh",
					ConfigFilePath: pathWorkspaceYaml,
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
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
			if err := ctrl.Update(ctx, logger, d.param); err != nil {
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
