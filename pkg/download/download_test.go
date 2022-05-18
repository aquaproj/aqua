package download_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	githubSvc "github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/flute/flute"
)

func stringP(s string) *string {
	return &s
}

func Test_pkgDownloader_GetReadCloser(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		param      *config.Param
		rt         *runtime.Runtime
		isErr      bool
		pkg        *config.Package
		pkgInfo    *config.PackageInfo
		assetName  string
		exp        string
		github     githubSvc.RepositoryService
		httpClient *http.Client
	}{
		{
			name: "github_release",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			pkg: &config.Package{
				Name:     "suzuki-shunsuke/ci-info",
				Registry: "standard",
				Version:  "v2.0.3",
			},
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
			},
			assetName: "ci-info-2.0.3_linux_amd64.tar.gz",
			exp:       "foo",
			github:    githubSvc.NewMock(nil, nil, "foo"),
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
											StatusCode: 200,
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
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			downloader := download.NewPackageDownloader(d.github, d.rt, download.NewHTTPDownloader(d.httpClient))
			file, err := downloader.GetReadCloser(ctx, d.pkg, d.pkgInfo, d.assetName, logE)
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
