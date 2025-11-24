package github

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/sirupsen/logrus"
)

type GHESRepoService struct {
	clients map[string]*github.RepositoriesService
}

func NewGHES(ctx context.Context, logE *logrus.Entry) *GHESRepoService {
	return &GHESRepoService{
		clients: make(map[string]*github.RepositoriesService),
	}
}

func (s *GHESRepoService) Resolve(baseURL string) (*RepositoriesService, error) {
	envKey, err := GetGitHubTokenEnvKey(baseURL)
	if err != nil {
		return nil, err
	}
	if envKey == TokenKeyGitHubCom {
		return New(context.Background(), logrus.NewEntry(logrus.New())), nil
	}
	if client, ok := s.clients[envKey]; ok {
		return client, nil
	}
	client, err := github.NewClient(MakeRetryable(
		getHTTPClientForGitHub(context.Background(), logrus.NewEntry(logrus.New()), getGitHubToken(envKey)), logrus.NewEntry(logrus.New()))).
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
