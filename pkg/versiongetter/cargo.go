package versiongetter

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

type CargoClient interface {
	ListVersions(ctx context.Context, crate string) ([]string, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
}

type CargoVersionGetter struct {
	client CargoClient
}

func NewCargo(client CargoClient) *CargoVersionGetter {
	return &CargoVersionGetter{
		client: client,
	}
}

func (c *CargoVersionGetter) Get(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	return c.client.GetLatestVersion(ctx, pkg.Crate) //nolint:wrapcheck
}

func (c *CargoVersionGetter) List(ctx context.Context, pkg *registry.PackageInfo, filters []*Filter) ([]*fuzzyfinder.Item, error) {
	versionStrings, err := c.client.ListVersions(ctx, pkg.Crate)
	if err != nil {
		return nil, fmt.Errorf("list versions of the crate: %w", err)
	}
	return fuzzyfinder.ConvertStringsToItems(versionStrings), nil
}
