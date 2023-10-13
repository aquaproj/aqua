package versiongetter

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

type VersionGetter interface {
	Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error)
	List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error)
}

type Generator struct {
	cargo     *CargoVersionGetter
	ghTag     *GitHubTagVersionGetter
	ghRelease *GitHubReleaseVersionGetter
}

func NewGenerator(cargo *CargoVersionGetter, ghTag *GitHubTagVersionGetter, ghRelease *GitHubReleaseVersionGetter) *Generator {
	return &Generator{
		cargo:     cargo,
		ghTag:     ghTag,
		ghRelease: ghRelease,
	}
}

func (g *Generator) Get(pkg *registry.PackageInfo) VersionGetter {
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
