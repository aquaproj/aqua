package versiongetter

import (
	"context"
	"errors"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

type MockGitHubTagClient struct {
	tags map[string][]*github.RepositoryTag
}

func NewMockGitHubTagClient(tags map[string][]*github.RepositoryTag) *MockGitHubTagClient {
	return &MockGitHubTagClient{
		tags: tags,
	}
}

func (g *MockGitHubTagClient) ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error) {
	tags, ok := g.tags[fmt.Sprintf("%s/%s", owner, repo)]
	if !ok {
		return nil, nil, errors.New("repository is not found")
	}
	m := (opts.Page + 1) * opts.PerPage
	if m > len(tags) {
		m = len(tags)
	}
	return tags[opts.Page*opts.PerPage : m], nil, nil
}
