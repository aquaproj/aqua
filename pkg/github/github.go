package github

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v44/github"
	"golang.org/x/oauth2"
)

type RepositoryService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

func New(httpClient *http.Client) RepositoryService {
	return github.NewClient(httpClient).Repositories
}

type AccessToken struct {
	token string
}

func NewAccessToken() *AccessToken {
	return &AccessToken{
		token: getGitHubToken(),
	}
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func NewHTTPClient(ctx context.Context, token *AccessToken) *http.Client {
	if token == nil || token.token == "" {
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.token},
	))
}
