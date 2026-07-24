package registry_test

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	cfgRegistry "github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/flute/flute"
)

// assertRegistryFilesArePrivate checks the permissions of the registry files
// downloaded into dir. Creating the file before writing it used to leave them
// at 0644, because WriteFile applies the permissions only when it creates the
// file itself.
//
// The JSON cache written next to each registry is skipped: createJSON uses
// os.Create and has always left it at 0644. That is inconsistent with the 0600
// of the registry it caches, but it is not what this change is about.
func assertRegistryFilesArePrivate(t *testing.T, dir string) {
	t.Helper()
	if err := filepath.WalkDir(dir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if filepath.Ext(p) == ".json" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err //nolint:wrapcheck
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("the permission of %s is %o, want 600", p, perm)
		}
		return nil
	}); err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestInstaller_InstallRegistries(t *testing.T) { //nolint:funlen
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	data := []struct {
		name        string
		files       map[string]string
		param       *config.Param
		downloader  registry.GitHubContentFileDownloader
		cfg         *aqua.Config
		cfgFilePath string
		isErr       bool
		exp         map[string]*cfgRegistry.Config
	}{
		{
			name: "local",
			param: &config.Param{
				MaxParallelism: 5,
			},
			cfgFilePath: "aqua.yaml",
			files: map[string]string{
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"local": {
						Type: "local",
						Name: "local",
						Path: "registry.yaml",
					},
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.16.0",
						Path:      "registry.yaml",
					},
					"standard-json": {
						Type:      "github_content",
						Name:      "standard-json",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.16.0",
						Path:      "registry.json",
					},
				},
			},
			exp: map[string]*cfgRegistry.Config{
				"local": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_content",
							RepoOwner: "aquaproj",
							RepoName:  "aqua-installer",
							Path:      "aqua-installer",
						},
					},
				},
				"standard": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "ci-info",
							Asset:     "ci-info_{{.Arch}}-{{.OS}}.tar.gz",
						},
					},
				},
				"standard-json": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "github-comment",
							Asset:     "github-comment_{{.Arch}}-{{.OS}}.tar.gz",
						},
					},
				},
			},
			downloader: download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://raw.githubusercontent.com",
							Routes: []flute.Route{
								{
									Name: "download a registry",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/aquaproj/aqua-registry/v2.16.0/registry.yaml",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: `packages:
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
`,
									},
								},
								{
									Name: "download a registry.json",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/aquaproj/aqua-registry/v2.16.0/registry.json",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: `{
  "packages": [
    {
      "type": "github_release",
      "repo_owner": "suzuki-shunsuke",
      "repo_name": "github-comment",
      "asset": "github-comment_{{.Arch}}-{{.OS}}.tar.gz"
    }
  ]
}
`,
									},
								},
							},
						},
					},
				},
			})),
		},
	}
	rt := &runtime.Runtime{
		GOOS:   "linux",
		GOARCH: "amd64",
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			// The registries are installed into the root directory, so it must
			// be a temporary directory rather than the working directory of the
			// test process.
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			d.param.RootDir = dir
			inst := registry.New(d.param, d.downloader, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			registries, err := inst.InstallRegistries(ctx, logger, d.cfg, filepath.Join(dir, d.cfgFilePath), nil)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, registries, cmp.AllowUnexported(cfgRegistry.Config{})); diff != "" {
				t.Fatal(diff)
			}
			assertRegistryFilesArePrivate(t, filepath.Join(dir, "registries"))
		})
	}
}
