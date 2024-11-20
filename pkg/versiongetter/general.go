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
	goGetter  *GoGetter
}

func NewGeneralVersionGetter(cargo *CargoVersionGetter, ghTag *GitHubTagVersionGetter, ghRelease *GitHubReleaseVersionGetter, goGetter *GoGetter) *GeneralVersionGetter {
	return &GeneralVersionGetter{
		cargo:     cargo,
		ghTag:     ghTag,
		ghRelease: ghRelease,
		goGetter:  goGetter,
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
	if pkg.GoVersionPath != "" {
		return g.goGetter
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
	if resp == nil {
		return
	}
	withRateLimitInfo(logE, resp).Debug("GitHub API Rate Limit info")
}

func withRateLimitInfo(logE *logrus.Entry, resp *github.Response) *logrus.Entry {
	if resp == nil {
		return logE
	}
	return logE.WithFields(logrus.Fields{
		"github_api_rate_limit":     resp.Rate.Limit,
		"github_api_rate_remaining": resp.Rate.Remaining,
	})
}
