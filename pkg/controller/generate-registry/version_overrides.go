package genrgst

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Package struct {
	Info    *registry.PackageInfo
	Version string
}

func getString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func getBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

type Release struct {
	ID      int64
	Tag     string
	Version *version.Version
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

func (ctrl *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logE *logrus.Entry, pkgName string, pkgInfo *registry.PackageInfo) (*registry.PackageInfo, []string) {
	ghReleases := ctrl.listReleases(ctx, logE, pkgInfo)
	releases := make([]*Release, len(ghReleases))
	for i, release := range ghReleases {
		tag := release.GetTagName()
		v, err := version.NewVersion(tag)
		if err != nil {
			logE.WithField("tag_name", tag).WithError(err).Warn("parse a tag as semver")
		}
		releases[i] = &Release{
			ID:      release.GetID(),
			Tag:     tag,
			Version: v,
		}
	}
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Version.GreaterThanOrEqual(releases[j].Version)
	})
	pkgs := make([]*Package, 0, len(releases))
	for _, release := range releases {
		pkgInfo := &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		assets := ctrl.listReleaseAssets(ctx, logE, pkgInfo, release.ID)
		logE.WithField("num_of_assets", len(assets)).Debug("got assets")
		if len(assets) == 0 {
			continue
		}
		ctrl.patchRelease(logE, pkgInfo, pkgName, release.Tag, assets)
		pkgs = append(pkgs, &Package{
			Info:    pkgInfo,
			Version: release.Tag,
		})
	}
	p, versions := mergePackages(pkgs)
	if p == nil {
		return pkgInfo, versions
	}
	p.Description = pkgInfo.Description
	p.Name = pkgInfo.Name
	return p, versions
}

func getVersionOverride(latestPkgInfo, pkgInfo *registry.PackageInfo) *registry.VersionOverride { //nolint:cyclop
	vo := &registry.VersionOverride{}
	if getString(pkgInfo.Asset) != getString(latestPkgInfo.Asset) {
		vo.Asset = pkgInfo.Asset
	}
	if pkgInfo.Format != latestPkgInfo.Format {
		vo.Format = pkgInfo.Format
	}
	if !reflect.DeepEqual(pkgInfo.Replacements, latestPkgInfo.Replacements) {
		vo.Replacements = pkgInfo.Replacements
		if pkgInfo.Replacements == nil {
			vo.Replacements = map[string]string{}
		}
	}
	if !reflect.DeepEqual(pkgInfo.Overrides, latestPkgInfo.Overrides) {
		vo.Overrides = pkgInfo.Overrides
		if pkgInfo.Overrides == nil {
			vo.Overrides = []*registry.Override{}
		}
	}
	if !reflect.DeepEqual(pkgInfo.SupportedEnvs, latestPkgInfo.SupportedEnvs) {
		vo.SupportedEnvs = pkgInfo.SupportedEnvs
		if pkgInfo.SupportedEnvs == nil {
			vo.SupportedEnvs = []string{}
		}
	}
	if getBool(pkgInfo.Rosetta2) != getBool(latestPkgInfo.Rosetta2) {
		vo.Rosetta2 = pkgInfo.Rosetta2
		if pkgInfo.Rosetta2 == nil {
			vo.Rosetta2 = boolP(false)
		}
	}
	if pkgInfo.WindowsExt != latestPkgInfo.WindowsExt {
		vo.WindowsExt = pkgInfo.WindowsExt
	}
	if !reflect.DeepEqual(pkgInfo.Checksum, latestPkgInfo.Checksum) {
		vo.Checksum = pkgInfo.Checksum
		if pkgInfo.Checksum == nil {
			vo.Checksum = &registry.Checksum{
				Enabled: boolP(false),
			}
		}
	}
	return vo
}

func mergePackages(pkgs []*Package) (*registry.PackageInfo, []string) {
	if len(pkgs) == 0 {
		return nil, nil
	}
	if len(pkgs) == 1 {
		return pkgs[0].Info, []string{pkgs[0].Version}
	}
	basePkg := pkgs[0]
	basePkgInfo := basePkg.Info
	latestPkgInfo := basePkgInfo
	latestVersion := basePkg.Version
	minimumVersion := basePkg.Version
	var lastMinimumVersion string
	vos := []*registry.VersionOverride{}
	var lastVO *registry.VersionOverride
	versions := []string{latestVersion}
	versionsM := map[string]struct{}{
		latestVersion: {},
	}
	for _, pkg := range pkgs[1:] {
		pkg := pkg
		pkgInfo := pkg.Info
		if reflect.DeepEqual(basePkgInfo, pkgInfo) {
			minimumVersion = pkg.Version
			continue
		}
		if _, ok := versionsM[minimumVersion]; !ok {
			versions = append(versions, minimumVersion)
			versionsM[minimumVersion] = struct{}{}
		}
		lastMinimumVersion = strings.TrimPrefix(minimumVersion, "v")
		if lastVO == nil {
			latestPkgInfo.VersionConstraints = fmt.Sprintf(`semver(">= %s")`, lastMinimumVersion)
		} else {
			lastVO.VersionConstraints = fmt.Sprintf(`semver(">= %s")`, lastMinimumVersion)
			vos = append(vos, lastVO)
		}
		lastVO = getVersionOverride(latestPkgInfo, pkgInfo)
		basePkgInfo = pkgInfo
		minimumVersion = pkg.Version
	}
	if lastMinimumVersion != "" {
		if _, ok := versionsM[minimumVersion]; !ok {
			versions = append(versions, minimumVersion)
			versionsM[minimumVersion] = struct{}{}
		}
		lastVO.VersionConstraints = fmt.Sprintf(`semver("< %s")`, lastMinimumVersion)
		vos = append(vos, lastVO)
	}
	if len(vos) != 0 {
		latestPkgInfo.VersionOverrides = vos
	}
	return latestPkgInfo, versions
}

func (ctrl *Controller) listReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryRelease {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var arr []*github.RepositoryRelease

	for i := 0; i < 10; i++ {
		releases, _, err := ctrl.github.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return arr
		}
		arr = append(arr, releases...)
		if len(releases) != opt.PerPage {
			return arr
		}
		opt.Page++
	}
	return arr
}
