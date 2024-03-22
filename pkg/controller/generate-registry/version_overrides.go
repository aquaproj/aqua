package genrgst

import (
	"context"
	"fmt"
	"sort"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
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

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logE *logrus.Entry, pkgName string, pkgInfo *registry.PackageInfo, limit int) (*registry.PackageInfo, []string) {
	ghReleases := c.listReleases(ctx, logE, pkgInfo, limit)
	releases := make([]*Release, len(ghReleases))
	for i, release := range ghReleases {
		tag := release.GetTagName()
		v, prefix, err := versiongetter.GetVersionAndPrefix(tag)
		if err != nil {
			logE.WithField("tag_name", tag).WithError(err).Warn("parse a tag as semver")
		}
		releases[i] = &Release{
			ID:            release.GetID(),
			Tag:           tag,
			Version:       v,
			VersionPrefix: prefix,
		}
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
			Type:      pkgInfo.Type,
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		if release.VersionPrefix != "" {
			pkgInfo.VersionPrefix = release.VersionPrefix
		}
		assets := c.listReleaseAssets(ctx, logE, pkgInfo, release.ID)
		logE.WithField("num_of_assets", len(assets)).Debug("got assets")
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
		PerPage: 100, //nolint:gomnd
	}

	if limit != 0 && limit < 100 {
		opt.PerPage = limit
	}

	var arr []*github.RepositoryRelease

	for i := 0; i < 10; i++ {
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
