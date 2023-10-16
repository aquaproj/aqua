package versiongetter

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

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

func (g *GeneralVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	getter := g.get(pkg)
	if getter == nil {
		return "", nil
	}
	return getter.Get(ctx, pkg, filters) //nolint:wrapcheck
}

func (g *GeneralVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	getter := g.get(pkg)
	if getter == nil {
		return nil, nil
	}
	return getter.List(ctx, pkg, filters) //nolint:wrapcheck
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
