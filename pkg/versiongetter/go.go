package versiongetter

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/hashicorp/go-version"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type GoGetter struct {
	gc GoProxyClient
}

func NewGoGetter(gc GoProxyClient) *GoGetter {
	return &GoGetter{
		gc: gc,
	}
}

type GoProxyClient interface {
	List(ctx context.Context, logger *slog.Logger, path string) ([]string, error)
}

func (g *GoGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter) (string, error) { //nolint:cyclop
	versions, err := g.gc.List(ctx, logger, pkg.GoVersionPath)
	if err != nil {
		return "", fmt.Errorf("list versions: %w", err)
	}
	var latest *version.Version
	var preLatest *version.Version
	for _, vs := range versions {
		v, err := version.NewSemver(vs)
		if err != nil {
			slogerr.WithError(logger, err).Warn("parse a version", "version", vs)
			continue
		}
		if v.Prerelease() == "" && (latest == nil || v.GreaterThan(latest)) {
			latest = v
			continue
		}
		if v.Prerelease() != "" && (preLatest == nil || v.GreaterThan(preLatest)) {
			preLatest = v
			continue
		}
	}
	if latest != nil {
		return latest.Original(), nil
	}
	if preLatest != nil {
		return preLatest.Original(), nil
	}
	return "", nil
}

func (g *GoGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, _ []*Filter, _ int) ([]*fuzzyfinder.Item, error) {
	vs, err := g.gc.List(ctx, logger, pkg.GoVersionPath)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	versions := make(version.Collection, 0, len(vs))
	for _, v := range vs {
		v, err := version.NewSemver(v)
		if err != nil {
			slogerr.WithError(logger, err).Warn("parse a version", "version", vs)
			continue
		}
		versions = append(versions, v)
	}
	sort.Sort(sort.Reverse(versions))
	items := make([]*fuzzyfinder.Item, len(versions))
	for i, version := range versions {
		items[i] = &fuzzyfinder.Item{
			Item: version.Original(),
		}
	}
	return items, nil
}
