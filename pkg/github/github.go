package github

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v83/github"
	"github.com/suzuki-shunsuke/ghtkn-go-sdk/ghtkn"
	"github.com/suzuki-shunsuke/go-retryablehttp"
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

func New(ctx context.Context, logger *slog.Logger) *RepositoriesService {
	return github.NewClient(MakeRetryable(getHTTPClientForGitHub(ctx, logger, getGitHubToken()), logger)).Repositories
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func MakeRetryable(client *http.Client, logger *slog.Logger) *http.Client {
	c := retryablehttp.NewClient()
	c.HTTPClient = client
	c.Logger = logger
	return c.StandardClient()
}

func getHTTPClientForGitHub(ctx context.Context, logger *slog.Logger, token string) *http.Client {
	if token == "" {
		if keyring.Enabled() {
			return oauth2.NewClient(ctx, ghtoken.NewTokenSource(logger, keyring.KeyService))
		}
		if os.Getenv("AQUA_GHTKN_ENABLED") == "true" {
			client := ghtkn.New()
			return oauth2.NewClient(ctx, client.TokenSource(logger, &ghtkn.InputGet{}))
		}
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}
