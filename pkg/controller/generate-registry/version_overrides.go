package genrgst

import (
	"context"
	"fmt"
	"sort"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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

func excludeVersion(logE *logrus.Entry, tag string, cfg *Config) bool {
	if cfg.Version == nil {
		return false
	}
	f, err := expr.EvaluateVersionFilter(cfg.Version, tag)
	if err != nil {
		logerr.WithError(logE, err).WithField("tag_name", tag).Warn("evaluate a version filter")
		return false
	}
	return !f
}

func excludeAsset(logE *logrus.Entry, asset string, cfg *Config) bool {
	if cfg.Asset == nil {
		return false
	}
	f, err := expr.EvaluateAssetFilter(cfg.Asset, asset)
	if err != nil {
		logerr.WithError(logE, err).WithField("asset", asset).Warn("evaluate an asset filter")
		return false
	}
	return !f
}

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logE *logrus.Entry, pkgName string, pkgInfo *registry.PackageInfo, limit int, cfg *Config) (*registry.PackageInfo, []string) { //nolint:cyclop,funlen
	ghReleases := c.listReleases(ctx, logE, pkgInfo, limit)
	releases := make([]*Release, 0, len(ghReleases))
	for _, release := range ghReleases {
		tag := release.GetTagName()
		if excludeVersion(logE, tag, cfg) {
			continue
		}
		v, prefix, err := versiongetter.GetVersionAndPrefix(tag)
		if err != nil {
			logerr.WithError(logE, err).WithField("tag_name", tag).Warn("parse a tag as semver")
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
		arr := c.listReleaseAssets(ctx, logE, pkgInfo, release.ID)
		logE.WithField("num_of_assets", len(arr)).Debug("got assets")
		assets := make([]*github.ReleaseAsset, 0, len(arr))
		for _, asset := range arr {
			if excludeAsset(logE, asset.GetName(), cfg) {
				continue
			}
			assets = append(assets, asset)
		}
		if len(assets) == 0 {
			continue
		}
		release.assets = assets
	}

	p, versions := c.generatePackage(logE, pkgName, releases)
	p.Type = pkgInfo.Type
	p.RepoOwner = pkgInfo.RepoOwner
	p.RepoName = pkgInfo.RepoName
	p.Description = pkgInfo.Description
	p.Name = pkgInfo.Name
	if len(p.VersionOverrides) != 0 {
		p.VersionConstraints = "false"
	}
	return p, versions
}

func (c *Controller) listReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo, limit int) []*github.RepositoryRelease {
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
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
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
