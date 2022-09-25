package download_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/flute/flute"
)

func stringP(s string) *string {
	return &s
}

func int64P(i int64) *int64 {
	return &i
}

func Test_PackageDownloader_GetReadCloser(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()
	data := []struct {
		name       string
		param      *config.Param
		rt         *runtime.Runtime
		isErr      bool
		pkg        *config.Package
		assetName  string
		exp        string
		github     domain.RepositoriesService
		httpClient *http.Client
	}{
		{ //nolint:dupl
			name: "github_release http",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "suzuki-shunsuke/ci-info",
					Registry: "standard",
					Version:  "v2.0.3",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
				},
			},
			assetName: "ci-info-2.0.3_linux_amd64.tar.gz",
			exp:       "foo",
			github: &github.MockRepositoriesService{
				Asset: "foo",
			},
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://github.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/suzuki-shunsuke/ci-info/releases/download/v2.0.3/ci-info-2.0.3_linux_amd64.tar.gz",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: "foo",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "github_release github api",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "suzuki-shunsuke/ci-info",
					Registry: "standard",
					Version:  "v2.0.3",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
				},
			},
			assetName: "ci-info-2.0.3_linux_amd64.tar.gz",
			exp:       "foo",
			github: &github.MockRepositoriesService{
				Releases: []*github.RepositoryRelease{
					{
						Assets: []*github.ReleaseAsset{
							{
								Name: stringP("ci-info-2.0.3_linux_amd64.tar.gz"),
								ID:   int64P(5),
							},
						},
					},
				},
				Asset: "foo",
			},
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://github.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/suzuki-shunsuke/ci-info/releases/download/v2.0.3/ci-info-2.0.3_linux_amd64.tar.gz",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusBadRequest,
										},
										BodyString: "invalid request",
									},
								},
							},
						},
					},
				},
			},
		},
		{ //nolint:dupl
			name: "github_content http",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "aquaproj/aqua-installer",
					Registry: "standard",
					Version:  "v1.1.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-installer",
					Path:      stringP("aqua-installer"),
				},
			},
			assetName: "aqua-installer",
			exp:       "foo",
			github: &github.MockRepositoriesService{
				Asset: "foo",
			},
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://raw.githubusercontent.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/aquaproj/aqua-installer/v1.1.0/aqua-installer",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: "foo",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "github_content http",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "aquaproj/aqua-installer",
					Registry: "standard",
					Version:  "v1.1.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-installer",
					Path:      stringP("aqua-installer"),
				},
			},
			assetName: "aqua-installer",
			exp:       "github-content",
			github: &github.MockRepositoriesService{
				Content: &github.RepositoryContent{
					Content: stringP("github-content"),
				},
				Asset: "foo",
			},
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://raw.githubusercontent.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/aquaproj/aqua-installer/v1.1.0/aqua-installer",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusBadRequest,
										},
										BodyString: "invalid request",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "github_archive",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "tfutils/tfenv",
					Registry: "standard",
					Version:  "v2.2.3",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
			},
			exp: "foo",
			github: &github.MockRepositoriesService{
				Asset: "foo",
			},
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://github.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/tfutils/tfenv/archive/refs/tags/v2.2.3.tar.gz",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: "foo",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "http",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:     "GoogleContainerTools/container-diff",
					Registry: "standard",
					Version:  "v0.17.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "http",
					RepoOwner: "GoogleContainerTools",
					RepoName:  "container-diff",
					URL:       stringP("https://storage.googleapis.com/container-diff/{{.Version}}/container-diff-{{.OS}}-amd64"),
				},
			},
			assetName: "container-diff-linux-amd64",
			exp:       "yoo",
			github:    nil,
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "https://storage.googleapis.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/container-diff/v0.17.0/container-diff-linux-amd64",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: http.StatusOK,
										},
										BodyString: "yoo",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid type",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: "invalid-type",
				},
			},
			isErr: true,
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			downloader := download.NewPackageDownloader(d.github, d.rt, download.NewHTTPDownloader(d.httpClient))
			file, _, err := downloader.GetReadCloser(ctx, d.pkg, d.assetName, logE)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			defer file.Close()
			b, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, string(b))
			}
		})
	}
}
