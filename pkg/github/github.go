package github

import (
	"context"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v72/github"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-retryablehttp-logrus/rlog"
	"golang.org/x/oauth2"
)

type (
	ReleaseAsset                = github.ReleaseAsset
	ListOptions                 = github.ListOptions
	RepositoryRelease           = github.RepositoryRelease
	RepositoriesService         = github.RepositoriesService
	Repository                  = github.Repository
	RepositoryContentGetOptions = github.RepositoryContentGetOptions
	RepositoryContent           = github.RepositoryContent
	Response                    = github.Response
	RepositoryTag               = github.RepositoryTag
	ArchiveFormat               = github.ArchiveFormat
)

const Tarball = github.Tarball

func New(ctx context.Context, logE *logrus.Entry) *RepositoriesService {
	return github.NewClient(retryHTTPClient(getHTTPClientForGitHub(ctx, logE, getGitHubToken()), logE)).Repositories
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func retryHTTPClient(client *http.Client, logE *logrus.Entry) *http.Client {
	c := retryablehttp.NewClient()
	c.HTTPClient = client
	c.Logger = rlog.New(logE)
	return c.StandardClient()
}

func getHTTPClientForGitHub(ctx context.Context, logE *logrus.Entry, token string) *http.Client {
	if token == "" {
		if keyring.Enabled() {
			return oauth2.NewClient(ctx, keyring.NewTokenSource(logE))
		}
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}
