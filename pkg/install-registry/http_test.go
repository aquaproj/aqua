package registry_test

import (
	"context"
	"net/http"
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
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/flute/flute"
)

func TestInstaller_HTTPRegistry(t *testing.T) { //nolint:funlen
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())
	data := []struct {
		name         string
		param        *config.Param
		cfg          *aqua.Config
		cfgFilePath  string
		isErr        bool
		exp          map[string]*cfgRegistry.Config
		roundTripper http.RoundTripper
	}{
		{
			name: "http registry - raw yaml",
			param: &config.Param{
				MaxParallelism: 5,
				RootDir:        "/tmp/aqua-root",
			},
			cfgFilePath: "aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"http-test": {
						Type:    "http",
						Name:    "http-test",
						URL:     "https://example.com/registry/{{.Version}}/registry.yaml",
						Version: "v1.0.0",
						Format:  "raw",
					},
				},
			},
			roundTripper: &flute.Transport{
				T: t,
				Services: []flute.Service{
					{
						Endpoint: "https://example.com",
						Routes: []flute.Route{
							{
								Tester: &flute.Tester{
									Method:       "GET",
									Path:         "/registry/v1.0.0/registry.yaml",
									PartOfHeader: http.Header{},
								},
								Response: &flute.Response{
									Base: http.Response{
										StatusCode: http.StatusOK,
									},
									BodyString: `packages:
- type: github_release
  repo_owner: cli
  repo_name: cli
  asset: gh_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz
`,
								},
							},
						},
					},
				},
			},
			exp: map[string]*cfgRegistry.Config{
				"http-test": {
					PackageInfos: cfgRegistry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "cli",
							RepoName:  "cli",
							Asset:     "gh_{{.Version}}_{{.OS}}_{{.Arch}}.tar.gz",
						},
					},
				},
			},
		},
		{
			name: "http registry - missing version template",
			param: &config.Param{
				MaxParallelism: 5,
				RootDir:        "/tmp/aqua-root",
			},
			cfgFilePath: "aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"http-invalid": {
						Type:    "http",
						Name:    "http-invalid",
						URL:     "https://example.com/registry/static/registry.yaml",
						Version: "v1.0.0",
					},
				},
			},
			isErr: true,
		},
		{
			name: "http registry - missing version",
			param: &config.Param{
				MaxParallelism: 5,
				RootDir:        "/tmp/aqua-root",
			},
			cfgFilePath: "aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"http-no-version": {
						Type: "http",
						Name: "http-no-version",
						URL:  "https://example.com/registry/{{.Version}}/registry.yaml",
					},
				},
			},
			isErr: true,
		},
		{
			name: "http registry - 404 error",
			param: &config.Param{
				MaxParallelism: 5,
				RootDir:        "/tmp/aqua-root",
			},
			cfgFilePath: "aqua.yaml",
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"http-404": {
						Type:    "http",
						Name:    "http-404",
						URL:     "https://example.com/registry/{{.Version}}/registry.yaml",
						Version: "v1.0.0",
					},
				},
			},
			roundTripper: &flute.Transport{
				T: t,
				Services: []flute.Service{
					{
						Endpoint: "https://example.com",
						Routes: []flute.Route{
							{
								Tester: &flute.Tester{
									Method:       "GET",
									Path:         "/registry/v1.0.0/registry.yaml",
									PartOfHeader: http.Header{},
								},
								Response: &flute.Response{
									Base: http.Response{
										StatusCode: http.StatusNotFound,
									},
									BodyString: "Not Found",
								},
							},
						},
					},
				},
			},
			isErr: true,
		},
	}
	rt := &runtime.Runtime{
		GOOS:   "linux",
		GOARCH: "amd64",
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			fs, err := testutil.NewFs(map[string]string{})
			if err != nil {
				t.Fatal(err)
			}

			var httpClient *http.Client
			if d.roundTripper != nil {
				httpClient = &http.Client{
					Transport: d.roundTripper,
				}
			} else {
				httpClient = http.DefaultClient
			}

			httpDownloader := download.NewHTTPDownloader(logE, httpClient)
			ghContentDownloader := download.NewGitHubContentFileDownloader(nil, httpDownloader)
			inst := registry.New(d.param, ghContentDownloader, httpDownloader, fs, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			registries, err := inst.InstallRegistries(ctx, logE, d.cfg, d.cfgFilePath, nil)
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
		})
	}
}
