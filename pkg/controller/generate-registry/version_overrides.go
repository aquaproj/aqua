package genrgst

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
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

func excludeVersion(logE *logrus.Entry, tag string, cfg *Config) bool {
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
			logerr.WithError(logE, err).WithField("tag_name", tag).Warn("evaluate a version filter")
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

func excludeAsset(logE *logrus.Entry, asset string, cfg *Config) bool {
	if cfg.AllAssetsFilter == nil {
		return false
	}
	f, err := expr.EvaluateAssetFilter(cfg.AllAssetsFilter, asset)
	if err != nil {
		logerr.WithError(logE, err).WithField("asset", asset).Warn("evaluate an asset filter")
		return false
	}
	return !f
}

func (c *Controller) getReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo, param *config.Param, cfg *Config) ([]*Release, error) {
	if param.AssetFile != "" {
		hrs := map[string][]string{}
		f, err := c.fs.Open(param.AssetFile)
		if err != nil {
			return nil, fmt.Errorf("open the asset file: %w", err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&hrs); err != nil {
			return nil, fmt.Errorf("read the asset file as JSON: %w", err)
		}
		return convHTTPReleases(logE, cfg, hrs), nil
	}

	ghReleases := c.listReleases(ctx, logE, pkgInfo, param.Limit)
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
	return releases, nil
}

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logE *logrus.Entry, pkgName string, pkgInfo *registry.PackageInfo, param *config.Param, cfg *Config) []string { //nolint:cyclop
	releases, err := c.getReleases(ctx, logE, pkgInfo, param, cfg)
	if err != nil {
		logerr.WithError(logE, err).WithField("pkg_name", pkgName).Error("get releases")
	}
	versions := c.generatePackage(logE, pkgInfo, pkgName, releases)
	if len(pkgInfo.VersionOverrides) != 0 {
		pkgInfo.VersionConstraints = "false"
	}
	return versions
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
