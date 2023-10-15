package versiongetter

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
)

type GitHubTagVersionGetter struct {
	gh GitHubTagClient
}

func NewGitHubTag(gh GitHubTagClient) *GitHubTagVersionGetter {
	return &GitHubTagVersionGetter{
		gh: gh,
	}
}

type GitHubTagClient interface {
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

func (g *GitHubTagVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	for {
		tags, _, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, tag := range tags {
			if filterTag(tag, filters) {
				return tag.GetName(), nil
			}
		}
		if len(tags) != opt.PerPage {
			return "", nil
		}
		opt.Page++
	}
}

func (g *GitHubTagVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	var versions []string
	tagNames := map[string]struct{}{}
	for {
		tags, _, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		for _, tag := range tags {
			tagName := tag.GetName()
			if _, ok := tagNames[tagName]; ok {
				continue
			}
			tagNames[tagName] = struct{}{}
			if filterTag(tag, filters) {
				versions = append(versions, tagName)
			}
		}
		if len(tags) != opt.PerPage {
			return fuzzyfinder.ConvertStringsToItems(versions), nil
		}
		opt.Page++
	}
}

func filterTag(tag *github.RepositoryTag, filters []*Filter) bool {
	tagName := tag.GetName()
	for _, filter := range filters {
		if filterTagByFilter(tagName, filter) {
			return true
		}
	}
	return false
}
