package genrgst

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Package struct {
	Info    *registry.PackageInfo
	Version string
	SemVer  string
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
	ID            int64
	Tag           string
	Version       *version.Version
	VersionPrefix string
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

var versionPattern = regexp.MustCompile(`^(.*?)v?((?:\d+)(?:\.\d+)?(?:\.\d+)?(?:(\.|-).+)?)$`)

func getVersionAndPrefix(tag string) (*version.Version, string, error) {
	if v, err := version.NewVersion(tag); err == nil {
		return v, "", nil
	}
	a := versionPattern.FindStringSubmatch(tag)
	if a == nil {
		return nil, "", nil
	}
	v, err := version.NewVersion(a[2])
	if err != nil {
		return nil, "", err //nolint:wrapcheck
	}
	return v, a[1], nil
}

func (c *Controller) getPackageInfoWithVersionOverrides(ctx context.Context, logE *logrus.Entry, pkgName string, pkgInfo *registry.PackageInfo) (*registry.PackageInfo, []string) {
	ghReleases := c.listReleases(ctx, logE, pkgInfo)
	releases := make([]*Release, len(ghReleases))
	for i, release := range ghReleases {
		tag := release.GetTagName()
		v, prefix, err := getVersionAndPrefix(tag)
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
			return r1.Tag >= r2.Tag
		}
		return v1.GreaterThanOrEqual(v2)
	})
	pkgs := make([]*Package, 0, len(releases))
	for _, release := range releases {
		pkgInfo := &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: pkgInfo.RepoOwner,
			RepoName:  pkgInfo.RepoName,
		}
		if release.VersionPrefix != "" {
			pkgInfo.VersionPrefix = &release.VersionPrefix
		}
		assets := c.listReleaseAssets(ctx, logE, pkgInfo, release.ID)
		logE.WithField("num_of_assets", len(assets)).Debug("got assets")
		if len(assets) == 0 {
			continue
		}
		c.patchRelease(logE, pkgInfo, pkgName, release.Tag, assets)
		pkgs = append(pkgs, &Package{
			Info:    pkgInfo,
			Version: release.Tag,
			SemVer:  release.Tag[len(release.VersionPrefix):],
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
			vo.Rosetta2 = util.BoolP(false)
		}
	}
	if pkgInfo.WindowsExt != latestPkgInfo.WindowsExt {
		vo.WindowsExt = pkgInfo.WindowsExt
	}
	if !reflect.DeepEqual(pkgInfo.VersionPrefix, latestPkgInfo.VersionPrefix) {
		vo.VersionPrefix = pkgInfo.VersionPrefix
		if pkgInfo.VersionPrefix == nil {
			vo.VersionPrefix = util.StrP("")
		}
	}
	if !reflect.DeepEqual(pkgInfo.Checksum, latestPkgInfo.Checksum) {
		vo.Checksum = pkgInfo.Checksum
		if pkgInfo.Checksum == nil {
			vo.Checksum = &registry.Checksum{
				Enabled: util.BoolP(false),
			}
		}
	}
	return vo
}

func isSemver(v string) bool {
	_, err := version.NewVersion(v)
	return err == nil
}

func mergePackages(pkgs []*Package) (*registry.PackageInfo, []string) { //nolint:funlen,cyclop
	if len(pkgs) == 0 {
		return nil, nil
	}
	if len(pkgs) == 1 {
		return pkgs[0].Info, []string{pkgs[0].Version}
	}
	basePkg := pkgs[0]
	basePkgInfo := basePkg.Info
	latestPkgInfo := basePkgInfo
	minimumVersion := basePkg.SemVer
	minimumTag := basePkg.Version
	var lastMinimumVersion string
	vos := []*registry.VersionOverride{}
	var lastVO *registry.VersionOverride
	versions := []string{basePkg.Version}
	versionsM := map[string]struct{}{
		basePkg.Version: {},
	}
	for _, pkg := range pkgs[1:] {
		pkgInfo := pkg.Info
		if reflect.DeepEqual(basePkgInfo, pkgInfo) {
			minimumVersion = pkg.SemVer
			minimumTag = pkg.Version
			continue
		}
		if _, ok := versionsM[minimumTag]; !ok {
			versions = append(versions, minimumTag)
			versionsM[minimumTag] = struct{}{}
		}
		lastMinimumVersion = strings.TrimPrefix(minimumVersion, "v")

		var versionConstraints string
		if isSemver(lastMinimumVersion) {
			versionConstraints = fmt.Sprintf(`semver(">= %s")`, lastMinimumVersion)
		} else {
			versionConstraints = fmt.Sprintf(`SemVer >= "%s"`, lastMinimumVersion)
		}
		if lastVO == nil {
			latestPkgInfo.VersionConstraints = versionConstraints
		} else {
			lastVO.VersionConstraints = versionConstraints
			vos = append(vos, lastVO)
		}
		lastVO = getVersionOverride(latestPkgInfo, pkgInfo)
		basePkgInfo = pkgInfo
		minimumVersion = pkg.SemVer
		minimumTag = pkg.Version
	}
	if lastMinimumVersion != "" {
		if _, ok := versionsM[minimumTag]; !ok {
			versions = append(versions, minimumTag)
			versionsM[minimumTag] = struct{}{}
		}
		if isSemver(lastMinimumVersion) {
			lastVO.VersionConstraints = fmt.Sprintf(`semver("< %s")`, lastMinimumVersion)
		} else {
			lastVO.VersionConstraints = fmt.Sprintf(`Version < "%s"`, lastMinimumVersion)
		}
		vos = append(vos, lastVO)
	}
	if len(vos) != 0 {
		latestPkgInfo.VersionOverrides = vos
	}
	return latestPkgInfo, versions
}

func (c *Controller) listReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryRelease {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var arr []*github.RepositoryRelease

	for i := 0; i < 10; i++ {
		releases, _, err := c.github.ListReleases(ctx, repoOwner, repoName, opt)
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
