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

func (g *GitHubTagVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: ghMaxPerPage, //nolint:gomnd
	}
	if limit > 0 && opt.PerPage > limit {
		opt.PerPage = limit
	}

	var versions []string
	tagNames := map[string]struct{}{}
	var prevCnt int // previous length of versions
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
				prevCnt = len(versions)
				versions = append(versions, tagName)
			}
		}
		if limit > 0 && len(versions) >= limit { // Reach the limit
			if len(versions) > limit {
				versions = versions[:limit]
			}
			return fuzzyfinder.ConvertStringsToItems(versions), nil
		}
		if len(tags) != opt.PerPage {
			return fuzzyfinder.ConvertStringsToItems(versions), nil
		}
		// After filtering, not enough versions added.
		// Increase per_page to reduce the consumption of GitHub API.
		if len(versions)-prevCnt < opt.PerPage && opt.PerPage < ghMaxPerPage {
			opt.PerPage *= 2
			if opt.PerPage > ghMaxPerPage {
				opt.PerPage = ghMaxPerPage
			}
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
