package which_test

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	cfgRegistry "github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

// rewrite roots the absolute paths of an expectation at dir. The test cases are
// written with readable paths such as /home/foo/aqua.yaml, but the files really
// live in a temporary directory. Relative paths, such as the path of a standard
// registry, are left alone.
func rewrite(dir string, r *which.FindResult) *which.FindResult {
	if r == nil {
		return nil
	}
	r.ExePath = testutil.Abs(dir, r.ExePath)
	r.ConfigFilePath = testutil.Abs(dir, r.ConfigFilePath)
	if r.Package != nil {
		if r.Package.Package != nil {
			r.Package.Package.FilePath = testutil.Abs(dir, r.Package.Package.FilePath)
		}
		if r.Package.Registry != nil {
			r.Package.Registry.Path = testutil.Abs(dir, r.Package.Registry.Path)
		}
	}
	if r.Config != nil {
		for _, pkg := range r.Config.Packages {
			pkg.FilePath = testutil.Abs(dir, pkg.FilePath)
		}
		for _, rgst := range r.Config.Registries {
			rgst.Path = testutil.Abs(dir, rgst.Path)
		}
	}
	return r
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
		exp     *which.FindResult
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:            pathHomeFooWorkspace,
				ConfigFilePath: "aqua.yaml",
				RootDir:        pathHomeFooLocalShare,
				MaxParallelism: 5,
			},
			exeName: pkgNameAquaInstaller,
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			exp: &which.FindResult{
				Package: &config.Package{
					Package: &aqua.Package{
						Name:     repoAquaInstaller,
						Registry: regTypeStandard,
						Version:  versionV1,
						FilePath: pathHomeFooWorkspaceAquaYaml,
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  pkgNameAquaInstaller,
						Path:      pkgNameAquaInstaller,
					},
					Registry: &aqua.Registry{
						Name: regTypeStandard,
						Type: regTypeLocal,
						Path: pathHomeFooWorkspaceRegistryYaml,
					},
				},
				File: &cfgRegistry.File{
					Name: pkgNameAquaInstaller,
				},
				Config: &aqua.Config{
					Packages: []*aqua.Package{
						{
							Name:     repoAquaInstaller,
							Registry: regTypeStandard,
							Version:  versionV1,
							FilePath: pathHomeFooWorkspaceAquaYaml,
						},
					},
					Registries: aqua.Registries{
						regTypeStandard: {
							Name: regTypeStandard,
							Type: regTypeLocal,
							Path: pathHomeFooWorkspaceRegistryYaml,
						},
					},
				},

				ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer",
				ConfigFilePath: pathHomeFooWorkspaceAquaYaml,
			},
		},
		{
			name: "outside aqua",
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:            pathHomeFooWorkspace,
				ConfigFilePath: "aqua.yaml",
				RootDir:        pathHomeFooLocalShare,
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/aquaproj-aqua/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: `packages:
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
			exp: &which.FindResult{
				ExePath: "/usr/local/bin/gh",
			},
		},
		{
			name: "global config",
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:                   pathHomeFooWorkspace,
				RootDir:               pathHomeFooLocalShare,
				MaxParallelism:        5,
				GlobalConfigFilePaths: []string{pathEtcAquaAquaYaml},
			},
			exeName: pkgNameAquaInstaller,
			files: map[string]string{
				pathEtcAquaAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: suzuki-shunsuke/ci-info@v1.0.0
- name: aquaproj/aqua-installer@v1.0.0
`,
				pathEtcAquaRegistryYaml: `packages:
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
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
			exp: &which.FindResult{
				Package: &config.Package{
					Package: &aqua.Package{
						Name:     repoAquaInstaller,
						Registry: regTypeStandard,
						Version:  versionV1,
						FilePath: pathEtcAquaAquaYaml,
					},
					PackageInfo: &cfgRegistry.PackageInfo{
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  pkgNameAquaInstaller,
						Path:      pkgNameAquaInstaller,
					},
					Registry: &aqua.Registry{
						Name: regTypeStandard,
						Type: regTypeLocal,
						Path: pathEtcAquaRegistryYaml,
					},
				},
				File: &cfgRegistry.File{
					Name: pkgNameAquaInstaller,
				},
				Config: &aqua.Config{
					Packages: []*aqua.Package{
						{
							Name:     "suzuki-shunsuke/ci-info",
							Registry: regTypeStandard,
							Version:  versionV1,
							FilePath: pathEtcAquaAquaYaml,
						},
						{
							Name:     repoAquaInstaller,
							Registry: regTypeStandard,
							Version:  versionV1,
							FilePath: pathEtcAquaAquaYaml,
						},
					},
					Registries: aqua.Registries{
						regTypeStandard: {
							Name: regTypeStandard,
							Type: regTypeLocal,
							Path: pathEtcAquaRegistryYaml,
						},
					},
				},
				ExePath:        "/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer",
				ConfigFilePath: pathEtcAquaAquaYaml,
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			linker := link.New()
			for dest, src := range d.links {
				src = testutil.Abs(dir, src)
				if err := osfile.MkdirAll(filepath.Dir(src)); err != nil {
					t.Fatal(err)
				}
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			testutil.RootParam(dir, d.param)
			env := testutil.RootEnv(dir, d.env)
			downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			ctrl := which.New(d.param, finder.NewConfigFinder(), reader.New(d.param), registry.New(d.param, downloader, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osenv.NewMock(env), linker)
			which, err := ctrl.Which(ctx, logger, d.param, d.exeName)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(rewrite(dir, d.exp), which); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
