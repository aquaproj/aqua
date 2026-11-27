package versiongetter

import (
	"context"
	"errors"
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
	files  VersionsFileReader
}

// VersionsFileReader reads a package's versions.json from aqua-registry-g2.
type VersionsFileReader interface {
	ReadVersionsFile(ctx context.Context, logger *slog.Logger, pkgName string) (*g2.Versions, error)
}

func NewG2(lister *g2.VersionLister, files VersionsFileReader) *G2VersionGetter {
	return &G2VersionGetter{lister: lister, files: files}
}

// Locked is the getter a project with a lock file chooses versions with: it reads the
// package's versions.json rather than its versions directory.
//
// The list is one small file rather than a tree of every version, and it says which
// versions upstream marked as prereleases, which the tags don't always say. It is a copy
// of the directory and may be behind it, so the newest version can be missing from it
// for a while; a version it does list is one the registry holds.
//
// Only a project with a lock file reads it. Choosing from it changes nothing there but
// where the list comes from, and a project without one, and aqua generate, keep reading
// the directory as they did.
func (g *G2VersionGetter) Locked() VersionGetter {
	return &lockedG2VersionGetter{g: g}
}

// Get returns the newest version.
func (g *G2VersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter) (string, error) {
	releases, err := g.releases(ctx, logger, pkg)
	if err != nil {
		return "", err
	}
	return latestTag(releases), nil
}

// List returns the versions, newest first.
func (g *G2VersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	releases, err := g.releases(ctx, logger, pkg)
	if err != nil {
		return nil, err
	}
	return listTags(releases, limit), nil
}

func latestTag(releases []*Release) string {
	latest := getLatestRelease(releases)
	if latest == nil {
		return ""
	}
	return latest.Tag
}

func listTags(releases []*Release, limit int) []*fuzzyfinder.Item {
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
	return fuzzyfinder.ConvertStringsToItems(versions)
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

// lockedG2VersionGetter chooses from the package's versions.json. See Locked.
type lockedG2VersionGetter struct {
	g *G2VersionGetter
}

func (l *lockedG2VersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter) (string, error) {
	releases, err := l.releases(ctx, logger, pkg)
	if err != nil {
		return "", err
	}
	return latestTag(releases), nil
}

func (l *lockedG2VersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	releases, err := l.releases(ctx, logger, pkg)
	if err != nil {
		return nil, err
	}
	return listTags(releases, limit), nil
}

// releases reads versions.json, and the versions directory when the file can't be read.
//
// The download can't tell a file that isn't there from one it failed to fetch, and a
// package with no version yet has no file. The directory can tell them apart: a package
// with no versions has none, and the tree says so with a 404. So the file is tried first
// and the directory decides the rest.
func (l *lockedG2VersionGetter) releases(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo) ([]*Release, error) {
	versions, err := l.g.files.ReadVersionsFile(ctx, logger, pkg.GetName())
	if err == nil {
		releases := make([]*Release, 0, len(versions.Versions))
		for _, v := range versions.Versions {
			if v == nil || v.Version == "" {
				continue
			}
			release := newRelease(v.Version)
			// The tag says it only when it is semver with a prerelease part. Upstream
			// can mark any release, and that mark is what the list carries.
			release.Prerelease = release.Prerelease || v.Prerelease
			releases = append(releases, release)
		}
		return releases, nil
	}
	if errors.Is(err, g2.ErrNoPackageBranch) {
		return nil, fmt.Errorf("read the versions of the package: %w", err)
	}
	logger.Debug("read the versions directory because versions.json couldn't be read", "error", err.Error())
	return l.g.releases(ctx, logger, pkg)
}
