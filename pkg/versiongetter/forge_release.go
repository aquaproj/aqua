package versiongetter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

// ForgeReleaseVersionGetter finds the versions of a package published as the releases of a
// repository on a Forgejo or Gitea instance.
type ForgeReleaseVersionGetter struct {
	client ForgeReleaseClient
}

// NewForgeRelease returns a version getter reading the given client.
func NewForgeRelease(client ForgeReleaseClient) *ForgeReleaseVersionGetter {
	return &ForgeReleaseVersionGetter{
		client: client,
	}
}

// ForgeReleaseClient lists the releases of a repository on a Forgejo or Gitea instance.
type ForgeReleaseClient interface {
	ListReleases(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Release, error)
}

func convForgeRelease(release *forge.Release) *Release {
	v, prefix, _ := GetVersionAndPrefix(release.TagName)
	return &Release{
		Tag:           release.TagName,
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    release.Prerelease || (v != nil && v.Prerelease() != ""),
	}
}

// filterForgeRelease says whether a release is one to install from. A draft is left out
// as well as a prerelease: its assets are not served to anyone but the people who can
// publish it.
func filterForgeRelease(logger *slog.Logger, release *forge.Release, filters []*Filter) bool {
	if release.Draft {
		return false
	}
	return filterTagName(logger, release.TagName, release.Prerelease, filters)
}

// Get returns the newest version the filters allow.
func (g *ForgeReleaseVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	if pkg.Host == "" {
		return "", errForgeHostRequired
	}
	candidates := []*Release{}
	// Pages are 1-based, and an empty page is the end of the list. A short page is not:
	// an instance can be configured to answer with fewer items than were asked for, and
	// stopping there would hide every release after the first page.
	for page := 1; page <= forge.MaxPages; page++ {
		releases, err := g.client.ListReleases(ctx, pkg.Host, pkg.RepoOwner, pkg.RepoName, page, forge.MaxPerPage)
		if err != nil {
			return "", fmt.Errorf("list releases: %w", err)
		}
		if len(releases) == 0 {
			return "", nil
		}
		for _, release := range releases {
			if filterForgeRelease(logger, release, filters) {
				candidates = append(candidates, convForgeRelease(release))
			}
		}
		if len(candidates) > 0 {
			return getLatestRelease(candidates).Tag, nil
		}
	}
	return "", nil
}

// List returns the versions the filters allow, newest first, for the version to be picked
// from.
func (g *ForgeReleaseVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	if pkg.Host == "" {
		return nil, errForgeHostRequired
	}
	perPage := forgeItemNumPerPage(limit, len(filters))
	var items []*fuzzyfinder.Item
	tags := map[string]struct{}{}
	for page := 1; page <= forge.MaxPages; page++ {
		releases, err := g.client.ListReleases(ctx, pkg.Host, pkg.RepoOwner, pkg.RepoName, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("list releases: %w", err)
		}
		if len(releases) == 0 {
			return items, nil
		}
		for _, release := range releases {
			if _, ok := tags[release.TagName]; ok {
				continue
			}
			tags[release.TagName] = struct{}{}
			if !filterForgeRelease(logger, release, filters) {
				continue
			}
			v := &fuzzyfinder.Version{
				Name:        release.Name,
				Version:     release.TagName,
				Description: release.Body,
				URL:         release.HTMLURL,
			}
			items = append(items, &fuzzyfinder.Item{
				Item:    release.TagName,
				Preview: fuzzyfinder.PreviewVersion(v),
			})
		}
		if limit > 0 && len(items) >= limit { // Reach the limit
			return items[:limit], nil
		}
	}
	return items, nil
}

// forgeItemNumPerPage asks for no more than the limit when every release asked for is
// one that can be shown, the same way itemNumPerPage does for GitHub. The cap is the
// instance's rather than GitHub's.
func forgeItemNumPerPage(limit, filterNum int) int {
	if limit > 0 && filterNum == 0 && forge.MaxPerPage > limit {
		return limit
	}
	return forge.MaxPerPage
}
