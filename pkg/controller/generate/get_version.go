package generate

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type VersionGetter interface {
	Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error)
	List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error)
}

type CargoClient interface {
	ListVersions(ctx context.Context, crate string) ([]string, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
}

type CargoVersionGetter struct {
	client CargoClient
}

func (c *CargoVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	return c.client.GetLatestVersion(ctx, pkg.Crate) //nolint:wrapcheck
}

func (c *CargoVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	versionStrings, err := c.client.ListVersions(ctx, pkg.Crate)
	if err != nil {
		return nil, fmt.Errorf("list versions of the crate: %w", err)
	}
	return fuzzyfinder.ConvertStringsToItems(versionStrings), nil
}

type GitHubTagVersionGetter struct {
	gh GitHubTagClient
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
	for {
		tags, _, err := g.gh.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		for _, tag := range tags {
			if filterTag(tag, filters) {
				versions = append(versions, tag.GetName())
			}
		}
		if len(tags) != opt.PerPage {
			return fuzzyfinder.ConvertStringsToItems(versions), nil
		}
		opt.Page++
	}
}

type GitHubReleaseVersionGetter struct {
	gh GitHubReleaseClient
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

// func (c *Controller) getVersionFromGitHub(ctx context.Context, logE *logrus.Entry, param *config.Param, pkgInfo *registry.PackageInfo) string {
// 	if pkgInfo.VersionSource == "github_tag" {
// 		return c.getVersionFromGitHubTag(ctx, logE, param, pkgInfo)
// 	}
// 	if param.SelectVersion {
// 		return c.selectVersionFromReleases(ctx, logE, pkgInfo)
// 	}
// 	if pkgInfo.VersionFilter != "" || pkgInfo.VersionPrefix != "" {
// 		return c.listAndGetTagName(ctx, logE, pkgInfo)
// 	}
// 	return c.getVersionFromLatestRelease(ctx, logE, pkgInfo)
// }

func (c *Controller) versionGetter(pkg *registry.PackageInfo) VersionGetter {
	if pkg.Type == "cargo" {
		return &CargoVersionGetter{
			client: c.cargoClient,
		}
	}
	if c.github == nil {
		return nil
	}
	if !pkg.HasRepo() {
		return nil
	}
	if pkg.VersionSource == "github_tag" {
		return &GitHubTagVersionGetter{
			gh: c.github,
		}
	}
	return &GitHubReleaseVersionGetter{
		gh: c.github,
	}
}

func (c *Controller) getVersion(ctx context.Context, _ *logrus.Entry, param *config.Param, pkg *fuzzyfinder.Package) string {
	if pkg.Version != "" {
		return pkg.Version
	}
	pkgInfo := pkg.PackageInfo

	filters, err := createFilters(pkgInfo)
	if err != nil {
		return ""
	}

	versionGetter := c.versionGetter(pkgInfo)
	if versionGetter == nil {
		return ""
	}

	if param.SelectVersion {
		versions, err := versionGetter.List(ctx, pkgInfo, filters)
		if err != nil {
			return ""
		}
		idx, err := c.fuzzyFinder.Find(versions, true)
		if err != nil {
			return ""
		}
		return versions[idx].Item
	}

	version, err := versionGetter.Get(ctx, pkgInfo, filters)
	if err != nil {
		return ""
	}
	return version
}
