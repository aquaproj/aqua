package generate

import (
	"strings"

	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/github"
)

// func (c *Controller) selectVersionFromReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
// 	releases := c.listReleases(ctx, logE, pkgInfo)
// 	items := make([]*fuzzyfinder.Item, len(releases))
// 	for i, release := range releases {
// 		items[i] = &fuzzyfinder.Item{
// 			Item: release.GetTagName(),
// 			Preview: fuzzyfinder.PreviewVersion(&fuzzyfinder.Version{
// 				Name:        release.GetName(),
// 				Version:     release.GetTagName(),
// 				Description: release.GetBody(),
// 				URL:         release.GetHTMLURL(),
// 			}),
// 		}
// 	}
// 	idx, err := c.fuzzyFinder.Find(items, true)
// 	if err != nil {
// 		return ""
// 	}
// 	return items[idx].Item
// }

// func (c *Controller) getVersionFromLatestRelease(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
// 	repoOwner := pkgInfo.RepoOwner
// 	repoName := pkgInfo.RepoName
// 	release, _, err := c.github.GetLatestRelease(ctx, repoOwner, repoName)
// 	if err != nil {
// 		logerr.WithError(logE, err).WithFields(logrus.Fields{
// 			"repo_owner": repoOwner,
// 			"repo_name":  repoName,
// 		}).Warn("get the latest release")
// 		return ""
// 	}
// 	return release.GetTagName()
// }

type Filter struct {
	Prefix     string
	Filter     *vm.Program
	Constraint string
}

func createFilters(pkgInfo *registry.PackageInfo) ([]*Filter, error) {
	filters := make([]*Filter, 0, 1+len(pkgInfo.VersionOverrides))
	topFilter := &Filter{}
	if pkgInfo.VersionFilter != "" {
		f, err := expr.CompileVersionFilter(pkgInfo.VersionFilter)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
		topFilter.Filter = f
	}
	topFilter.Constraint = pkgInfo.VersionConstraints
	if pkgInfo.VersionPrefix != "" {
		topFilter.Prefix = pkgInfo.VersionPrefix
	}
	filters = append(filters, topFilter)

	for _, vo := range pkgInfo.VersionOverrides {
		flt := &Filter{
			Prefix:     topFilter.Prefix,
			Filter:     topFilter.Filter,
			Constraint: topFilter.Constraint,
		}
		if vo.VersionFilter != nil {
			f, err := expr.CompileVersionFilter(*vo.VersionFilter)
			if err != nil {
				return nil, err //nolint:wrapcheck
			}
			flt.Filter = f
		}
		flt.Constraint = vo.VersionConstraints
		if vo.VersionPrefix != nil {
			flt.Prefix = *vo.VersionPrefix
		}
		filters = append(filters, flt)
	}
	return filters, nil
}

// func (c *Controller) listReleases(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryRelease {
// 	repoOwner := pkgInfo.RepoOwner
// 	repoName := pkgInfo.RepoName
// 	opt := &github.ListOptions{
// 		PerPage: 100, //nolint:gomnd
// 	}
// 	var arr []*github.RepositoryRelease
//
// 	filters, err := createFilters(pkgInfo)
// 	if err != nil {
// 		return nil
// 	}
//
// 	for i := 0; i < 10; i++ {
// 		releases, _, err := c.github.ListReleases(ctx, repoOwner, repoName, opt)
// 		if err != nil {
// 			logerr.WithError(logE, err).WithFields(logrus.Fields{
// 				"repo_owner": repoOwner,
// 				"repo_name":  repoName,
// 			}).Warn("list releases")
// 			return arr
// 		}
// 		for _, release := range releases {
// 			if filterRelease(release, filters) {
// 				arr = append(arr, release)
// 			}
// 		}
// 		if len(releases) != opt.PerPage {
// 			return arr
// 		}
// 		opt.Page++
// 	}
// 	return arr
// }

// func (c *Controller) listAndGetTagName(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
// 	repoOwner := pkgInfo.RepoOwner
// 	repoName := pkgInfo.RepoName
// 	opt := &github.ListOptions{
// 		PerPage: 30, //nolint:gomnd
// 	}
//
// 	filters, err := createFilters(pkgInfo)
// 	if err != nil {
// 		return ""
// 	}
//
// 	for {
// 		releases, _, err := c.github.ListReleases(ctx, repoOwner, repoName, opt)
// 		if err != nil {
// 			logerr.WithError(logE, err).WithFields(logrus.Fields{
// 				"repo_owner": repoOwner,
// 				"repo_name":  repoName,
// 			}).Warn("list releases")
// 			return ""
// 		}
// 		for _, release := range releases {
// 			if filterRelease(release, filters) {
// 				return release.GetTagName()
// 			}
// 		}
// 		if len(releases) != opt.PerPage {
// 			return ""
// 		}
// 		opt.Page++
// 	}
// }

func filterRelease(release *github.RepositoryRelease, filters []*Filter) bool {
	if release.GetPrerelease() {
		return false
	}

	tagName := release.GetTagName()

	for _, filter := range filters {
		if filterTagByFilter(tagName, filter) {
			return true
		}
	}
	return false
}

func filterTagByFilter(tagName string, filter *Filter) bool {
	sv := tagName
	if filter.Prefix != "" {
		if !strings.HasPrefix(tagName, filter.Prefix) {
			return false
		}
		sv = strings.TrimPrefix(tagName, filter.Prefix)
	}
	if filter.Filter != nil {
		if f, err := expr.EvaluateVersionFilter(filter.Filter, tagName); err != nil || !f {
			return false
		}
	}
	if filter.Constraint == "" {
		return true
	}
	if f, err := expr.EvaluateVersionConstraints(filter.Constraint, tagName, sv); err == nil && f {
		return true
	}
	return false
}
