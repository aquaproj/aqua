package versiongetter

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
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
	List(ctx context.Context, path string) ([]string, error)
}

func (g *GoGetter) Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	versions, err := g.gc.List(ctx, pkg.GoVersionPath)
	if err != nil {
		return "", fmt.Errorf("list versions: %w", err)
	}
	var latest *version.Version
	for _, vs := range versions {
		v, err := version.NewSemver(vs)
		if err != nil {
			logE.WithError(err).WithField("version", vs).Warn("parse a version")
			continue
		}
		if latest == nil || v.GreaterThan(latest) {
			latest = v
		}
	}
	return latest.String(), nil
}

func (g *GoGetter) List(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, _ []*Filter, _ int) ([]*fuzzyfinder.Item, error) {
	versions, err := g.gc.List(ctx, pkg.GoVersionPath)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	items := make([]*fuzzyfinder.Item, len(versions))
	for i, version := range versions {
		items[i] = &fuzzyfinder.Item{
			Item: version,
		}
	}
	return items, nil
}
