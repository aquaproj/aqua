package versiongetter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forgejo"
)

// ForgejoReleaseVersionGetter finds the versions of a package published as the releases
// of a repository on a Forgejo instance.
type ForgejoReleaseVersionGetter struct {
	client ForgejoReleaseClient
}

// NewForgejoRelease returns a version getter reading the given client.
func NewForgejoRelease(client ForgejoReleaseClient) *ForgejoReleaseVersionGetter {
	return &ForgejoReleaseVersionGetter{
		client: client,
	}
}

// ForgejoReleaseClient lists the releases of a repository on a Forgejo instance.
type ForgejoReleaseClient interface {
	ListReleases(ctx context.Context, host, owner, repo string, page, limit int) ([]*forgejo.Release, error)
}

func convForgejoRelease(release *forgejo.Release) *Release {
	v, prefix, _ := GetVersionAndPrefix(release.TagName)
	return &Release{
		Tag:           release.TagName,
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    release.Prerelease || (v != nil && v.Prerelease() != ""),
	}
}

// filterForgejoRelease says whether a release is one to install from. A draft is left out
// as well as a prerelease: its assets are not served to anyone but the people who can
// publish it.
func filterForgejoRelease(logger *slog.Logger, release *forgejo.Release, filters []*Filter) bool {
	if release.Draft {
		return false
	}
	return filterTagName(logger, release.TagName, release.Prerelease, filters)
}

// Get returns the newest version the filters allow.
func (g *ForgejoReleaseVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	if pkg.Host == "" {
		return "", errForgejoHostRequired
	}
	candidates := []*Release{}
	// Pages are 1-based, and an empty page is the end of the list. A short page is not:
	// an instance can be configured to answer with fewer items than were asked for, and
	// stopping there would hide every release after the first page.
	for page := 1; page <= forgejo.MaxPages; page++ {
		releases, err := g.client.ListReleases(ctx, pkg.Host, pkg.RepoOwner, pkg.RepoName, page, forgejo.MaxPerPage)
		if err != nil {
			return "", fmt.Errorf("list releases: %w", err)
		}
		if len(releases) == 0 {
			return "", nil
		}
		for _, release := range releases {
			if filterForgejoRelease(logger, release, filters) {
				candidates = append(candidates, convForgejoRelease(release))
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
func (g *ForgejoReleaseVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	if pkg.Host == "" {
		return nil, errForgejoHostRequired
	}
	perPage := forgejoItemNumPerPage(limit, len(filters))
	var items []*fuzzyfinder.Item
	tags := map[string]struct{}{}
	for page := 1; page <= forgejo.MaxPages; page++ {
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
			if !filterForgejoRelease(logger, release, filters) {
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

// forgejoItemNumPerPage asks for no more than the limit when every release asked for is
// one that can be shown, the same way itemNumPerPage does for GitHub. The cap is the
// instance's rather than GitHub's.
func forgejoItemNumPerPage(limit, filterNum int) int {
	if limit > 0 && filterNum == 0 && forgejo.MaxPerPage > limit {
		return limit
	}
	return forgejo.MaxPerPage
}
