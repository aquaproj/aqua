package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/google/go-github/v93/github"
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

func New(ctx context.Context, logger *slog.Logger) (*RepositoriesService, error) {
	httpClient, err := getHTTPClientForGitHub(ctx, logger, getGitHubToken())
	if err != nil {
		return nil, err
	}
	client, err := github.NewClient(github.WithHTTPClient(MakeRetryable(httpClient, logger)))
	if err != nil {
		return nil, fmt.Errorf("create a GitHub client: %w", err)
	}
	return client.Repositories, nil
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

func getHTTPClientForGitHub(ctx context.Context, logger *slog.Logger, token string) (*http.Client, error) {
	if token != "" {
		return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)), nil
	}
	ts, err := NewTokenSource(logger)
	if err != nil {
		return nil, err
	}
	if ts == nil {
		return http.DefaultClient, nil
	}
	return oauth2.NewClient(ctx, ts), nil
}

// NewTokenSource returns a token source getting a GitHub access token from the keyring or ghtkn.
// It doesn't read environment variables such as GITHUB_TOKEN.
// The keyring takes precedence over ghtkn.
// It returns nil if neither the keyring nor ghtkn is enabled.
func NewTokenSource(logger *slog.Logger) (oauth2.TokenSource, error) {
	if ts, err := NewKeyringTokenSource(logger); ts != nil || err != nil {
		return ts, err
	}
	return NewGhtknTokenSource(logger)
}

// NewKeyringTokenSource returns a token source getting a GitHub access token from the keyring.
// It returns nil if the keyring isn't enabled.
func NewKeyringTokenSource(logger *slog.Logger) (oauth2.TokenSource, error) {
	if !keyring.Enabled() {
		return nil, nil //nolint:nilnil
	}
	return ghtoken.NewTokenSource(logger, keyring.KeyService), nil
}

// NewGhtknTokenSource returns a token source getting a GitHub access token from ghtkn.
// It returns nil if ghtkn isn't enabled.
func NewGhtknTokenSource(logger *slog.Logger) (oauth2.TokenSource, error) {
	ghtknEnabled, err := ghtkn.Enabled(&ghtkn.InputEnabled{
		Envs: []string{
			"AQUA_GHTKN_ENABLED",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("check if ghtkn is enabled: %w", err)
	}
	if !ghtknEnabled {
		return nil, nil //nolint:nilnil
	}
	client, err := ghtkn.New()
	if err != nil {
		return nil, fmt.Errorf("create a ghtkn client: %w", err)
	}
	return client.TokenSource(logger, &ghtkn.InputGet{}), nil
}
