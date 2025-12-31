package download_test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/suzuki-shunsuke/flute/flute"
)

func TestGitHubContentFileDownloader_DownloadGitHubContentFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		param      *domain.GitHubContentFileParam
		github     download.GitHub
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
				Content: "foo",
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
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			downloader := download.NewGitHubContentFileDownloader(d.github, download.NewHTTPDownloader(logger, d.httpClient))
			file, err := downloader.DownloadGitHubContentFile(ctx, logger, d.param)
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
