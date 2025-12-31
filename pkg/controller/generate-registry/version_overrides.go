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
	assets        []*github.ReleaseAsset
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
		f, err := expr.EvaluateVersionFilter(cfg.VersionFilter, tag)
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

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logger *slog.Logger, pkgName string, pkgInfo *registry.PackageInfo, limit int, cfg *Config) []string { //nolint:cyclop
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
		releases = append(releases, &Release{
			ID:            release.GetID(),
			Tag:           tag,
			Version:       v,
			VersionPrefix: prefix,
		})
	}
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
	for _, release := range releases {
		pkgInfo := &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		if release.VersionPrefix != "" {
			pkgInfo.VersionPrefix = release.VersionPrefix
		}
		arr := c.listReleaseAssets(ctx, logger, pkgInfo, release.ID)
		logger.Debug("got assets", "num_of_assets", len(arr))
		assets := make([]*github.ReleaseAsset, 0, len(arr))
		for _, asset := range arr {
			if excludeAsset(logger, asset.GetName(), cfg) {
				continue
			}
			assets = append(assets, asset)
		}
		if len(assets) == 0 {
			continue
		}
		release.assets = assets
	}

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
			slogerr.WithError(logger, err).Warn("list releases",
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
