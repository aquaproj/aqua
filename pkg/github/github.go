package github

import (
	"context"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v80/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/ghtkn-go-sdk/ghtkn"
	"github.com/suzuki-shunsuke/go-retryablehttp"
	"github.com/suzuki-shunsuke/go-retryablehttp-logrus/rlog"
	"github.com/suzuki-shunsuke/slog-logrus/slogrus"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/keyring/ghtoken"
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

func New(ctx context.Context, logE *logrus.Entry, ts oauth2.TokenSource, httpClient *http.Client) *RepositoriesService {
	if hc := getHTTPClientForGitHub(ctx, ts); hc != nil {
		httpClient = hc
	}
	return github.NewClient(MakeRetryable(httpClient, logE)).Repositories
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func MakeRetryable(client *http.Client, logE *logrus.Entry) *http.Client {
	c := retryablehttp.NewClient()
	c.HTTPClient = client
	c.Logger = rlog.New(logE)
	return c.StandardClient()
}

func getHTTPClientForGitHub(ctx context.Context, ts oauth2.TokenSource) *http.Client {
	if ts == nil {
		return nil
	}
	return oauth2.NewClient(ctx, ts)
}

func NewTokenSource(logE *logrus.Entry) oauth2.TokenSource {
	if token := getGitHubToken(); token != "" {
		return oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
	}
	if keyring.Enabled() {
		return ghtoken.NewTokenSource(logE, keyring.KeyService)
	}
	if os.Getenv("AQUA_GHTKN_ENABLED") == "true" {
		client := ghtkn.New()
		return client.TokenSource(slogrus.Convert(logE), &ghtkn.InputGet{})
	}
	return nil
}
