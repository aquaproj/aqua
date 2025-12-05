package github

import (
	"context"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v79/github"
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

const (
	Tarball           = github.Tarball
	TokenKeyGitHubCom = "GITHUB_TOKEN" //nolint:gosec
	TokenKeyGHES      = "GITHUB_ENTERPRISE_TOKEN"
)

func New(ctx context.Context, logE *logrus.Entry) *RepositoriesService {
	return github.NewClient(MakeRetryable(getHTTPClientForGitHub(ctx, logE, getGitHubToken(TokenKeyGitHubCom)), logE)).Repositories
}

func getGitHubToken(envKey string) string {
	if token := os.Getenv("AQUA_" + envKey); token != "" {
		return token
	}
	return os.Getenv(envKey)
}

func MakeRetryable(client *http.Client, logE *logrus.Entry) *http.Client {
	c := retryablehttp.NewClient()
	c.HTTPClient = client
	c.Logger = rlog.New(logE)
	return c.StandardClient()
}

func getHTTPClientForGitHub(ctx context.Context, logE *logrus.Entry, token string) *http.Client {
	if token == "" {
		if keyring.Enabled() {
			return oauth2.NewClient(ctx, ghtoken.NewTokenSource(logE, keyring.KeyService))
		}
		if os.Getenv("AQUA_GHTKN_ENABLED") == "true" {
			client := ghtkn.New()
			return oauth2.NewClient(ctx, client.TokenSource(slogrus.Convert(logE), &ghtkn.InputGet{}))
		}
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}
