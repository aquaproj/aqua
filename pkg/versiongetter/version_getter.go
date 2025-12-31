package versiongetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

type VersionGetter interface {
	Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error)
	List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error)
}
