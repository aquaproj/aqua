package versiongetter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
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

func (g *GitHubReleaseVersionGetter) Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName

	var respToLog *github.Response
	defer func() {
		logGHRateLimit(logE, respToLog)
	}()

	release, resp, err := g.gh.GetLatestRelease(ctx, repoOwner, repoName)
	respToLog = resp
	if err != nil {
		return "", fmt.Errorf("get the latest GitHub Release: %w", err)
	}

	if len(filters) == 0 || filterRelease(release, filters) {
		return release.GetTagName(), nil
	}

	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	for {
		releases, resp, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
		respToLog = resp
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, release := range releases {
			if filterRelease(release, filters) {
				return release.GetTagName(), nil
			}
		}
		if resp.NextPage == 0 {
			return "", nil
		}
		opt.Page = resp.NextPage
	}
}

func (g *GitHubReleaseVersionGetter) List(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: itemNumPerPage(limit, len(filters)),
	}

	var respToLog *github.Response
	defer func() {
		addRteLimitInfo(logE, respToLog)
	}()

	var items []*fuzzyfinder.Item
	tags := map[string]struct{}{}
	for {
		releases, resp, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
		respToLog = resp
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		for _, release := range releases {
			tagName := release.GetTagName()
			if _, ok := tags[tagName]; ok {
				continue
			}
			tags[tagName] = struct{}{}
			if filterRelease(release, filters) {
				v := &fuzzyfinder.Version{
					Name:        release.GetName(),
					Version:     tagName,
					Description: release.GetBody(),
					URL:         release.GetHTMLURL(),
				}
				items = append(items, &fuzzyfinder.Item{
					Item:    tagName,
					Preview: fuzzyfinder.PreviewVersion(v),
				})
			}
		}
		if limit > 0 && len(items) >= limit { // Reach the limit
			return items[:limit], nil
		}
		if resp.NextPage == 0 {
			return items, nil
		}
		opt.Page = resp.NextPage
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
