package versiongetter

import (
	"context"
	"errors"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

type MockVersionGetter struct {
	versions map[string][]*fuzzyfinder.Item
}

func NewMockVersionGetter(versions map[string][]*fuzzyfinder.Item) *MockVersionGetter {
	return &MockVersionGetter{
		versions: versions,
	}
}

func (g *MockVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	versions, ok := g.versions[pkg.GetName()]
	if !ok {
		return "", errors.New("version isn't found")
	}
	return versions[0].Item, nil
}

func (g *MockVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	versions, ok := g.versions[pkg.GetName()]
	if !ok {
		return nil, errors.New("version isn't found")
	}
	return versions, nil
}
