package versiongetter

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

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

func (g *GitHubTagVersionGetter) Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	for {
		tags, resp, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logGHRateLimit(logE, resp)
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, tag := range tags {
			if filterTag(tag, filters) {
				logGHRateLimit(logE, resp)
				return tag.GetName(), nil
			}
		}
		if resp.NextPage == 0 {
			logGHRateLimit(logE, resp)
			return "", nil
		}
		opt.Page = resp.NextPage
	}
}

func (g *GitHubTagVersionGetter) List(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: itemNumPerPage(limit, len(filters)),
	}

	var versions []string
	tagNames := map[string]struct{}{}
	for {
		tags, resp, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		*logE = *logE.WithFields(logrus.Fields{ // finder's output will overwrite the log, add fields to parent logE here
			"gh_rate_limit":     resp.Rate.Limit,
			"gh_rate_remaining": resp.Rate.Remaining,
		})
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
		if limit > 0 && len(versions) >= limit { // Reach the limit
			return fuzzyfinder.ConvertStringsToItems(versions[:limit]), nil
		}
		if resp.NextPage == 0 {
			return fuzzyfinder.ConvertStringsToItems(versions), nil
		}
		opt.Page = resp.NextPage
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
