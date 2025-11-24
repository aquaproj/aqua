package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/sirupsen/logrus"
)

type GitHub interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, maxRedirects int) (*url.URL, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
}

type GHESRepositoryService struct {
	clients map[string]*github.RepositoriesService
}

func NewGHES(standard *github.RepositoriesService) *GHESRepositoryService {
	clients := make(map[string]*github.RepositoriesService)
	clients[TokenKeyGitHubCom] = standard
	return &GHESRepositoryService{
		clients: clients,
	}
}

func (s *GHESRepositoryService) Resolve(ctx context.Context, logE *logrus.Entry, baseURL string) (GitHub, error) {
	envKey, err := GetGitHubTokenEnvKey(baseURL)
	if err != nil {
		return nil, err
	}
	if client, ok := s.clients[envKey]; ok {
		return client, nil
	}
	client, err := github.NewClient(MakeRetryable(
		getHTTPClientForGitHub(ctx, logE, getGitHubToken(envKey)), logrus.NewEntry(logrus.New()))).
		WithEnterpriseURLs(baseURL, "")
	if err != nil {
		return nil, err
	}
	s.clients[envKey] = client.Repositories
	return client.Repositories, nil
}

func GetGitHubTokenEnvKey(baseURL string) (string, error) {
	if baseURL == "" {
		return TokenKeyGitHubCom, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	// extract domain
	d := strings.TrimSpace(u.Host)
	if d == "" {
		return "", errors.New("invalid domain")
	}

	d = strings.ToLower(d)
	if !regexp.MustCompile(`^[a-z0-9.-]+\.[a-z0-9]+$`).MatchString(d) {
		return "", errors.New("invalid domain")
	}

	if d == "github.com" {
		return TokenKeyGitHubCom, nil
	}

	transformed := strings.ReplaceAll(d, ".", "_")
	return "GITHUB_TOKEN_" + transformed, nil
}
