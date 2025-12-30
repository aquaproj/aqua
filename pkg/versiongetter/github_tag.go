package versiongetter

import (
	"context"
	"fmt"
	"log/slog"

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

func convTag(tag *github.RepositoryTag) *Release {
	v, prefix, _ := GetVersionAndPrefix(tag.GetName())
	return &Release{
		Tag:           tag.GetName(),
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    v.Prerelease() != "",
	}
}

func (g *GitHubTagVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:mnd
	}

	var respToLog *github.Response
	defer func() {
		logGHRateLimit(logger, respToLog)
	}()

	candidates := []*Release{}

	for {
		tags, resp, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		respToLog = resp
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, tag := range tags {
			if filterTag(tag, filters) {
				candidates = append(candidates, convTag(tag))
			}
		}
		if len(candidates) > 0 {
			return getLatestRelease(candidates).Tag, nil
		}
		if resp.NextPage == 0 {
			return "", nil
		}
		opt.Page = resp.NextPage
	}
}

func (g *GitHubTagVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: itemNumPerPage(limit, len(filters)),
	}

	var versions []string
	tagNames := map[string]struct{}{}
	for {
		tags, resp, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logger = withRateLimitInfo(logger, resp)
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
		if matchTagByFilter(tagName, filter) {
			return !filter.NoAsset
		}
	}
	return false
}
