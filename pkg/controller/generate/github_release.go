package generate

import (
	"context"

	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/expr"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) selectVersionFromReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
	releases := ctrl.listReleases(ctx, logE, pkgInfo)
	versions := make([]*Version, len(releases))
	for i, release := range releases {
		versions[i] = &Version{
			Name:        release.GetName(),
			Version:     release.GetTagName(),
			Description: release.GetBody(),
			URL:         release.GetHTMLURL(),
		}
	}
	idx, err := ctrl.versionSelector.Find(versions)
	if err != nil {
		return ""
	}
	return versions[idx].Version
}

func (ctrl *Controller) getVersionFromLatestRelease(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	release, _, err := ctrl.github.GetLatestRelease(ctx, repoOwner, repoName)
	if err != nil {
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"repo_owner": repoOwner,
			"repo_name":  repoName,
		}).Warn("get the latest release")
		return ""
	}
	return release.GetTagName()
}

func (ctrl *Controller) listReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryRelease { //nolint:cyclop
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var versionFilter *vm.Program
	if pkgInfo.VersionFilter != nil {
		var err error
		versionFilter, err = expr.CompileVersionFilter(*pkgInfo.VersionFilter)
		if err != nil {
			return nil
		}
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
		for _, release := range releases {
			if release.GetPrerelease() {
				continue
			}
			if versionFilter != nil {
				f, err := expr.EvaluateVersionFilter(versionFilter, release.GetTagName())
				if err != nil || !f {
					continue
				}
			}
			arr = append(arr, release)
		}
		if len(releases) != opt.PerPage {
			return arr
		}
		opt.Page++
	}
	return arr
}

func (ctrl *Controller) listAndGetTagName(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	versionFilter, err := expr.CompileVersionFilter(*pkgInfo.VersionFilter)
	if err != nil {
		return ""
	}
	for {
		releases, _, err := ctrl.github.ListReleases(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return ""
		}
		for _, release := range releases {
			if release.GetPrerelease() {
				continue
			}
			f, err := expr.EvaluateVersionFilter(versionFilter, release.GetTagName())
			if err != nil || !f {
				continue
			}
			return release.GetTagName()
		}
		if len(releases) != opt.PerPage {
			return ""
		}
		opt.Page++
	}
}
