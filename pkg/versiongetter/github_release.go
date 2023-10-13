package versiongetter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
)

type GitHubReleaseVersionGetter struct {
	gh GitHubReleaseClient
}

func NewGitHubRelease(gh GitHubReleaseClient) *GitHubReleaseVersionGetter {
	return &GitHubReleaseVersionGetter{
		gh: gh,
	}
}

type GitHubReleaseClient interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

func (g *GitHubReleaseVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName

	if len(filters) == 0 {
		release, _, err := g.gh.GetLatestRelease(ctx, repoOwner, repoName)
		if err != nil {
			return "", fmt.Errorf("get the latest GitHub Release: %w", err)
		}
		return release.GetTagName(), nil
	}

	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	for {
		releases, _, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, release := range releases {
			if filterRelease(release, filters) {
				return release.GetTagName(), nil
			}
		}
		if len(releases) != opt.PerPage {
			return "", nil
		}
		opt.Page++
	}
}

func (g *GitHubReleaseVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	var items []*fuzzyfinder.Item
	for {
		releases, _, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		for _, release := range releases {
			if filterRelease(release, filters) {
				v := &fuzzyfinder.Version{
					Name:        release.GetName(),
					Version:     release.GetTagName(),
					Description: release.GetBody(),
					URL:         release.GetHTMLURL(),
				}
				items = append(items, &fuzzyfinder.Item{
					Item:    release.GetName(),
					Preview: fuzzyfinder.PreviewVersion(v),
				})
			}
		}
		if len(releases) != opt.PerPage {
			return items, nil
		}
		opt.Page++
	}
}

func filterRelease(release *github.RepositoryRelease, filters []*Filter) bool {
	if release.GetPrerelease() {
		return false
	}

	tagName := release.GetTagName()

	for _, filter := range filters {
		if filterTagByFilter(tagName, filter) {
			return true
		}
	}
	return false
}

func filterTagByFilter(tagName string, filter *Filter) bool {
	sv := tagName
	if filter.Prefix != "" {
		if !strings.HasPrefix(tagName, filter.Prefix) {
			return false
		}
		sv = strings.TrimPrefix(tagName, filter.Prefix)
	}
	if filter.Filter != nil {
		if f, err := expr.EvaluateVersionFilter(filter.Filter, tagName); err != nil || !f {
			return false
		}
	}
	if filter.Constraint == "" {
		return true
	}
	if f, err := expr.EvaluateVersionConstraints(filter.Constraint, tagName, sv); err == nil && f {
		return true
	}
	return false
}
