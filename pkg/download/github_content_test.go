package download_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/flute/flute"
)

func stringP(s string) *string {
	return &s
}

func TestGitHubContentFileDownloader_DownloadGitHubContentFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		param      *domain.GitHubContentFileParam
		github     github.RepositoriesService
		httpClient *http.Client
		isErr      bool
		exp        string
	}{
		{
			name: "github_content http",
			param: &domain.GitHubContentFileParam{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Ref:       "v2.16.0",
				Path:      "registry.yaml",
			},
			exp:    "foo",
			github: nil,
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
										Path:   "/aquaproj/aqua-registry/v2.16.0/registry.yaml",
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
			name: "github_content github api",
			param: &domain.GitHubContentFileParam{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Ref:       "v2.16.0",
				Path:      "registry.yaml",
			},
			exp: "foo",
			github: &github.MockRepositoriesService{
				Content: &github.RepositoryContent{
					Content: stringP("foo"),
				},
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
										Path:   "/aquaproj/aqua-registry/v2.16.0/registry.yaml",
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
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			downloader := download.NewGitHubContentFileDownloader(d.github, download.NewHTTPDownloader(d.httpClient))
			file, err := downloader.DownloadGitHubContentFile(ctx, logE, d.param)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if file.String != "" {
				if file.String != d.exp {
					t.Fatalf("wanted %s, got %s", d.exp, file.String)
				}
				return
			}
			defer file.ReadCloser.Close()
			if d.isErr {
				t.Fatal("error must be returned")
			}
			b, err := io.ReadAll(file.ReadCloser)
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
