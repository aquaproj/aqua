package versiongetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
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

func (g *GeneralVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	getter := g.get(pkg)
	if getter == nil {
		return "", nil
	}
	return getter.Get(ctx, logger, pkg, filters) //nolint:wrapcheck
}

func (g *GeneralVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	getter := g.get(pkg)
	if getter == nil {
		return nil, nil
	}
	return getter.List(ctx, logger, pkg, filters, limit) //nolint:wrapcheck
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
func logGHRateLimit(logger *slog.Logger, resp *github.Response) {
	if resp == nil {
		return
	}
	withRateLimitInfo(logger, resp).Debug("GitHub API Rate Limit info")
}

func withRateLimitInfo(logger *slog.Logger, resp *github.Response) *slog.Logger {
	if resp == nil {
		return logger
	}
	return logger.With(
		"github_api_rate_limit", resp.Rate.Limit,
		"github_api_rate_remaining", resp.Rate.Remaining,
	)
}
