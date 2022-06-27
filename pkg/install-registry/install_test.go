package registry_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/config/clivm"
	cfgRegistry "github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/download"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/flute/flute"
)

func stringP(s string) *string {
	return &s
}

func Test_installer_InstallRegistries(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name        string
		files       map[string]string
		param       *config.Param
		downloader  download.RegistryDownloader
		cfg         *clivm.Config
		cfgFilePath string
		isErr       bool
		exp         map[string]*cfgRegistry.Config
	}{
		{
			name: "local",
			param: &config.Param{
				MaxParallelism: 5,
			},
			cfgFilePath: "clivm.yaml",
			files: map[string]string{
				"registry.yaml": `packages:
- type: github_content
  repo_owner: clivm
  repo_name: clivm-installer
  path: clivm-installer
`,
			},
			cfg: &clivm.Config{
				Registries: clivm.Registries{
					"local": {
						Type: "local",
						Name: "local",
						Path: "registry.yaml",
					},
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						RepoOwner: "clivm",
						RepoName:  "clivm-registry",
						Ref:       "v2.16.0",
						Path:      "registry.yaml",
					},
					"standard-json": {
						Type:      "github_content",
						Name:      "standard-json",
						RepoOwner: "clivm",
						RepoName:  "clivm-registry",
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
							RepoOwner: "clivm",
							RepoName:  "clivm-installer",
							Path:      stringP("clivm-installer"),
						},
					},
				},
				"standard": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{.Arch}}-{{.OS}}.tar.gz"),
						},
					},
				},
				"standard-json": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "github-comment",
							Asset:     stringP("github-comment_{{.Arch}}-{{.OS}}.tar.gz"),
						},
					},
				},
			},
			downloader: download.NewRegistryDownloader(nil, download.NewHTTPDownloader(&http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://raw.githubusercontent.com",
							Routes: []flute.Route{
								{
									Name: "download a registry",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/clivm/clivm-registry/v2.16.0/registry.yaml",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: 200,
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
										Path:   "/clivm/clivm-registry/v2.16.0/registry.json",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: 200,
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
			inst := registry.New(d.param, d.downloader, fs)
			registries, err := inst.InstallRegistries(ctx, d.cfg, d.cfgFilePath, logE)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, registries); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
