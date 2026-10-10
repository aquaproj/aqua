package genrgst

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/hashicorp/go-version"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type Package struct {
	Info    *registry.PackageInfo
	Version string
	SemVer  string
}

type Release struct {
	ID            int64
	Tag           string
	Version       *version.Version
	VersionPrefix string
	// assets are the names of the files the release publishes, which is all that is
	// read of them: what tells two releases apart is how they name what they publish.
	assets []string
}

func (r *Release) LessThan(r2 *Release) bool {
	if r.Version != nil && r2.Version != nil {
		return r.Version.LessThan(r2.Version)
	}
	return r.Tag < r2.Tag
}

func (r *Release) GreaterThan(r2 *Release) bool {
	if r.Version != nil && r2.Version != nil {
		return r.Version.GreaterThan(r2.Version)
	}
	return r.Tag > r2.Tag
}

func listPkgsFromVersions(pkgName string, versions []string) []*aqua.Package {
	if len(versions) == 0 {
		return nil
	}
	pkgs := []*aqua.Package{
		{
			Name: fmt.Sprintf("%s@%s", pkgName, versions[0]),
		},
	}
	for _, v := range versions[1:] {
		pkgs = append(pkgs, &aqua.Package{
			Name:    pkgName,
			Version: v,
		})
	}
	return pkgs
}

func excludeVersion(logger *slog.Logger, tag string, cfg *Config) bool {
	excludedVersions := map[string]struct{}{
		"latest":  {},
		"nightly": {},
		"stable":  {},
	}
	if _, ok := excludedVersions[tag]; ok {
		return true
	}
	if cfg.VersionFilter != nil {
		f, err := expr.EvaluateVersionFilter(logger, cfg.VersionFilter, tag)
		if err != nil {
			slogerr.WithError(logger, err).Warn("evaluate a version filter", "tag_name", tag)
			return false
		}
		if !f {
			return true
		}
	}
	if cfg.VersionPrefix != "" {
		if !strings.HasPrefix(tag, cfg.VersionPrefix) {
			return true
		}
	}
	return false
}

func excludeAsset(logger *slog.Logger, asset string, cfg *Config) bool {
	if cfg.AllAssetsFilter == nil {
		return false
	}
	f, err := expr.EvaluateAssetFilter(cfg.AllAssetsFilter, asset)
	if err != nil {
		slogerr.WithError(logger, err).Warn("evaluate an asset filter", "asset", asset)
		return false
	}
	return !f
}

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logger *slog.Logger, pkgName string, pkgInfo *registry.PackageInfo, limit int, cfg *Config) []string {
	return c.versionOverrides(logger, pkgName, pkgInfo, c.gitHubReleases(ctx, logger, pkgInfo, limit, cfg))
}

// gitHubReleases are the releases of a GitHub repository, with what each publishes.
//
// The asset list comes per release rather than with it, because a release carries only
// the first page of its own.
func (c *Controller) gitHubReleases(ctx context.Context, logger *slog.Logger, pkgInfo *registry.PackageInfo, limit int, cfg *Config) []*Release {
	ghReleases := c.listReleases(ctx, logger, pkgInfo, limit)
	releases := make([]*Release, 0, len(ghReleases))
	for _, release := range ghReleases {
		tag := release.GetTagName()
		if excludeVersion(logger, tag, cfg) {
			continue
		}
		v, prefix, err := versiongetter.GetVersionAndPrefix(tag)
		if err != nil {
			slogerr.WithError(logger, err).Warn("parse a tag as semver", "tag_name", tag)
		}
		rel := &Release{
			ID:            release.GetID(),
			Tag:           tag,
			Version:       v,
			VersionPrefix: prefix,
		}
		info := &registry.PackageInfo{
			Type:      pkgTypeGitHubRelease,
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		arr := c.listReleaseAssets(ctx, logger, info, rel.ID)
		logger.Debug("got assets", "num_of_assets", len(arr))
		for _, asset := range arr {
			if excludeAsset(logger, asset.GetName(), cfg) {
				continue
			}
			rel.assets = append(rel.assets, asset.GetName())
		}
		releases = append(releases, rel)
	}
	return releases
}

// versionOverrides builds the definition that answers for every release it is given, which
// is what reading more than one of them is for: a package whose asset naming changed is a
// version_override saying so.
func (c *Controller) versionOverrides(logger *slog.Logger, pkgName string, pkgInfo *registry.PackageInfo, releases []*Release) []string {
	sort.Slice(releases, func(i, j int) bool {
		r1 := releases[i]
		r2 := releases[j]
		v1 := r1.Version
		v2 := r2.Version
		if v1 == nil || v2 == nil {
			return r1.Tag <= r2.Tag
		}
		return v1.LessThan(v2)
	})
	versions := c.generatePackage(logger, pkgInfo, pkgName, releases)
	if len(pkgInfo.VersionOverrides) != 0 {
		pkgInfo.VersionConstraints = "false"
	}
	return versions
}

func (c *Controller) listReleases(ctx context.Context, logger *slog.Logger, pkgInfo *registry.PackageInfo, limit int) []*github.RepositoryRelease {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:mnd
	}

	if limit != 0 && limit < 100 {
		opt.PerPage = limit
	}

	var arr []*github.RepositoryRelease

	for range 10 {
		releases, resp, err := c.github.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			slogerr.WithError(logger, err).Warn(
				"list releases",
				"repo_owner", repoOwner,
				"repo_name", repoName,
			)
			return arr
		}
		arr = append(arr, releases...)
		if limit > 0 && len(releases) >= limit {
			return arr
		}
		if resp.NextPage == 0 {
			return arr
		}
		opt.Page = resp.NextPage
	}
	return arr
}
