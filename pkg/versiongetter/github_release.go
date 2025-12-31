package versiongetter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/hashicorp/go-version"
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

type Release struct {
	Tag           string
	Version       *version.Version
	VersionPrefix string
	Prerelease    bool
}

func convRelease(release *github.RepositoryRelease) *Release {
	v, prefix, _ := GetVersionAndPrefix(release.GetTagName())
	return &Release{
		Tag:           release.GetTagName(),
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    release.GetPrerelease() || (v != nil && v.Prerelease() != ""),
	}
}

func compareRelease(latest, release *Release) bool {
	if latest.Prerelease && !release.Prerelease {
		return true
	}
	if !latest.Prerelease && release.Prerelease {
		return false
	}
	if release.Version == nil {
		if latest.Version != nil {
			return false
		}
		if release.Tag > latest.Tag {
			return true
		}
		return false
	}
	if latest.Version == nil {
		return true
	}
	if release.Version.GreaterThan(latest.Version) {
		return true
	}
	return false
}

func getLatestRelease(releases []*Release) *Release {
	if len(releases) == 0 {
		return nil
	}
	latest := releases[0]
	for _, release := range releases[1:] {
		if compareRelease(latest, release) {
			latest = release
			continue
		}
	}
	return latest
}

func (g *GitHubReleaseVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName

	var respToLog *github.Response
	defer func() {
		logGHRateLimit(logger, respToLog)
	}()

	candidates := []*Release{}

	opt := &github.ListOptions{
		PerPage: 30, //nolint:mnd
	}
	for {
		releases, resp, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
		respToLog = resp
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}
		for _, release := range releases {
			if filterRelease(release, filters) {
				candidates = append(candidates, convRelease(release))
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

func (g *GitHubReleaseVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	repoOwner := pkg.RepoOwner
	repoName := pkg.RepoName
	opt := &github.ListOptions{
		PerPage: itemNumPerPage(limit, len(filters)),
	}

	var items []*fuzzyfinder.Item
	tags := map[string]struct{}{}
	for {
		releases, resp, err := g.gh.ListReleases(ctx, repoOwner, repoName, opt)
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
		if matchTagByFilter(tagName, filter) {
			return !filter.NoAsset
		}
	}
	return false
}
