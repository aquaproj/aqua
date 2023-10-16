package versiongetter

import (
	"context"
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

type MockGitHubReleaseClient struct {
	releases map[string][]*github.RepositoryRelease
}

func NewMockGitHubReleaseClient(releases map[string][]*github.RepositoryRelease) *MockGitHubReleaseClient {
	return &MockGitHubReleaseClient{
		releases: releases,
	}
}

func (g *MockGitHubReleaseClient) GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error) {
	releases, ok := g.releases[fmt.Sprintf("%s/%s", repoOwner, repoName)]
	if !ok {
		return nil, nil, errors.New("repository isn't found")
	}
	return releases[0], nil, nil
}

func (g *MockGitHubReleaseClient) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error) {
	releases, ok := g.releases[fmt.Sprintf("%s/%s", owner, repo)]
	if !ok {
		return nil, nil, errors.New("repository isn't found")
	}
	return releases, nil, nil
}
