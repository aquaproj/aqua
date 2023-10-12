package generate

import (
	"github.com/aquaproj/aqua/v2/pkg/github"
)

func filterTag(tag *github.RepositoryTag, filters []*Filter) bool {
	tagName := tag.GetName()
	for _, filter := range filters {
		if filterTagByFilter(tagName, filter) {
			return true
		}
	}
	return false
}

// listTags lists GitHub Tags by GitHub API and filter them with `version_filter`.
// func (c *Controller) listTags(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryTag {
// 	// List GitHub Tags by GitHub API
// 	// Filter tags with version_filter
// 	repoOwner := pkgInfo.RepoOwner
// 	repoName := pkgInfo.RepoName
// 	opt := &github.ListOptions{
// 		PerPage: 100, //nolint:gomnd
// 	}
//
// 	filters, err := createFilters(pkgInfo)
// 	if err != nil {
// 		return nil
// 	}
//
// 	var arr []*github.RepositoryTag
// 	for i := 0; i < 10; i++ {
// 		tags, _, err := c.github.ListTags(ctx, repoOwner, repoName, opt)
// 		if err != nil {
// 			logerr.WithError(logE, err).WithFields(logrus.Fields{
// 				"repo_owner": repoOwner,
// 				"repo_name":  repoName,
// 			}).Warn("list releases")
// 			return arr
// 		}
// 		for _, tag := range tags {
// 			if filterTag(tag, filters) {
// 				arr = append(arr, tag)
// 			}
// 		}
// 		if len(tags) != opt.PerPage {
// 			return arr
// 		}
// 		opt.Page++
// 	}
// 	return arr
// }

// func (c *Controller) listAndGetTagNameFromTag(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
// 	// List GitHub Tags by GitHub API
// 	// Filter tags with version_filter
// 	// Get a tag
// 	repoOwner := pkgInfo.RepoOwner
// 	repoName := pkgInfo.RepoName
// 	opt := &github.ListOptions{
// 		PerPage: 30, //nolint:gomnd
// 	}
// 	filters, err := createFilters(pkgInfo)
// 	if err != nil {
// 		return ""
// 	}
// 	for {
// 		tags, _, err := c.github.ListTags(ctx, repoOwner, repoName, opt)
// 		if err != nil {
// 			logerr.WithError(logE, err).WithFields(logrus.Fields{
// 				"repo_owner": repoOwner,
// 				"repo_name":  repoName,
// 			}).Warn("list tags")
// 			return ""
// 		}
// 		for _, tag := range tags {
// 			if filterTag(tag, filters) {
// 				return tag.GetName()
// 			}
// 		}
// 		if len(tags) != opt.PerPage {
// 			return ""
// 		}
// 		opt.Page++
// 	}
// }

// func (c *Controller) selectVersionFromGitHubTag(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
// 	tags := c.listTags(ctx, logE, pkgInfo)
// 	items := make([]*fuzzyfinder.Item, len(tags))
// 	for i, tag := range tags {
// 		items[i] = &fuzzyfinder.Item{
// 			Item: tag.GetName(),
// 		}
// 	}
// 	idx, err := c.fuzzyFinder.Find(items, false)
// 	if err != nil {
// 		return ""
// 	}
// 	return items[idx].Item
// }

// func (c *Controller) getVersionFromGitHubTag(ctx context.Context, logE *logrus.Entry, param *config.Param, pkgInfo *registry.PackageInfo) string {
// 	if param.SelectVersion {
// 		return c.selectVersionFromGitHubTag(ctx, logE, pkgInfo)
// 	}
// 	return c.listAndGetTagNameFromTag(ctx, logE, pkgInfo)
// }
