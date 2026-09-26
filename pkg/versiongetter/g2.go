package versiongetter

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/g2"
)

// G2VersionGetter offers the versions aqua-registry-g2 has generated a registry.json
// for, rather than every version upstream has released.
//
// Those are the versions aqua can lock, so offering any other one would only lead to
// an "aqua lock update" that fails. It also means the choice no longer depends on a
// registry: the package branch is the list.
type G2VersionGetter struct {
	lister *g2.VersionLister
}

func NewG2(lister *g2.VersionLister) *G2VersionGetter {
	return &G2VersionGetter{lister: lister}
}

// Get returns the newest version.
func (g *G2VersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter) (string, error) {
	releases, err := g.releases(ctx, logger, pkg)
	if err != nil {
		return "", err
	}
	latest := getLatestRelease(releases)
	if latest == nil {
		return "", nil
	}
	return latest.Tag, nil
}

// List returns the versions, newest first.
func (g *G2VersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	releases, err := g.releases(ctx, logger, pkg)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(releases, func(a, b *Release) int {
		switch {
		case compareRelease(a, b):
			return 1
		case compareRelease(b, a):
			return -1
		default:
			return 0
		}
	})
	versions := make([]string, len(releases))
	for i, release := range releases {
		versions[i] = release.Tag
	}
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}
	return fuzzyfinder.ConvertStringsToItems(versions), nil
}

// releases reads the package branch.
//
// The filters are ignored. version_filter and version_prefix are applied by ar2 when
// it generates registry.json, so a version being on the branch already means it
// passed them; applying them again here would only be a second chance to disagree.
func (g *G2VersionGetter) releases(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo) ([]*Release, error) {
	versions, err := g.lister.List(ctx, logger, pkg.GetName())
	if err != nil {
		return nil, fmt.Errorf("list the versions of the package: %w", err)
	}
	releases := make([]*Release, len(versions))
	for i, v := range versions {
		releases[i] = newRelease(v)
	}
	return releases, nil
}

// newRelease reads a version out of a tag.
func newRelease(tag string) *Release {
	v, prefix, _ := GetVersionAndPrefix(tag)
	return &Release{
		Tag:           tag,
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    v != nil && v.Prerelease() != "",
	}
}
