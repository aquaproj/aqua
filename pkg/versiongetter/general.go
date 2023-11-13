package versiongetter

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

const ghMaxPerPage int = 100

type GeneralVersionGetter struct {
	cargo     *CargoVersionGetter
	ghTag     *GitHubTagVersionGetter
	ghRelease *GitHubReleaseVersionGetter
}

func NewGeneralVersionGetter(cargo *CargoVersionGetter, ghTag *GitHubTagVersionGetter, ghRelease *GitHubReleaseVersionGetter) *GeneralVersionGetter {
	return &GeneralVersionGetter{
		cargo:     cargo,
		ghTag:     ghTag,
		ghRelease: ghRelease,
	}
}

func (g *GeneralVersionGetter) Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	getter := g.get(pkg)
	if getter == nil {
		return "", nil
	}
	return getter.Get(ctx, logE, pkg, filters) //nolint:wrapcheck
}

func (g *GeneralVersionGetter) List(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	getter := g.get(pkg)
	if getter == nil {
		return nil, nil
	}
	return getter.List(ctx, logE, pkg, filters, limit) //nolint:wrapcheck
}

func (g *GeneralVersionGetter) get(pkg *registry.PackageInfo) VersionGetter {
	if pkg.Type == "cargo" {
		return g.cargo
	}
	if g.ghTag == nil {
		return nil
	}
	if !pkg.HasRepo() {
		return nil
	}
	if pkg.VersionSource == "github_tag" {
		return g.ghTag
	}
	return g.ghRelease
}

// If filters exist, filter as much data as possible and
// per_page would be ghMaxPerPage.
// If there are no filters, set per_page to the limit
// to reduce the size of response data.
func itemNumPerPage(limit, filterNum int) int {
	if limit > 0 && filterNum == 0 && ghMaxPerPage > limit {
		return limit
	}
	return ghMaxPerPage
}

// log the GitHub API rate limit info
func logGHRateLimit(logE *logrus.Entry, resp *github.Response) {
	logE.WithFields(logrus.Fields{
		"gh_rate_limit":     resp.Rate.Limit,
		"gh_rate_remaining": resp.Rate.Remaining,
	}).Info("GitHub API rate limit info")
}

// fuzzy-finder's output will overwrite the log, add fields to original logE
func addRteLimitInfo(logE *logrus.Entry, resp *github.Response) {
	*logE = *logE.WithFields(logrus.Fields{
		"gh_rate_limit":     resp.Rate.Limit,
		"gh_rate_remaining": resp.Rate.Remaining,
	})
}
