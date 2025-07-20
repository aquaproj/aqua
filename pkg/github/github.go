package github

import (
	"context"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v73/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-retryablehttp"
	"github.com/suzuki-shunsuke/go-retryablehttp-logrus/rlog"
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

// GetUserAgent returns a Chrome browser User-Agent string for better compatibility with cloud storage services
// This is a workaround until the origin issue is resolved.
// See: https://github.com/aquaproj/aqua/pull/4019#issuecomment-3092666269
func GetUserAgent() string {
	agent := os.Getenv("AQUA_DOWNLOAD_USER_AGENT")
	if agent != "" {
		return agent
	}
	return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"
}

func New(ctx context.Context, logE *logrus.Entry) *RepositoriesService {
	return github.NewClient(MakeRetryable(getHTTPClientForGitHub(ctx, logE, getGitHubToken()), logE)).Repositories
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

func getHTTPClientForGitHub(ctx context.Context, logE *logrus.Entry, token string) *http.Client {
	if token == "" {
		if keyring.Enabled() {
			return oauth2.NewClient(ctx, ghtoken.NewTokenSource(logE, keyring.KeyService))
		}
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}
