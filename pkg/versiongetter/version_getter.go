package versiongetter

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
)

// API: Get versions from data sources like GitHub Releases, GitHub Tags, Go Modules, etc.
// Filter: Filter versions
// Formatter: Format versions
// Sorter: Sort versions
// Picker: Pick out a version
// FuzzyFinder: Find a version by fuzzy search

type VersionGetter interface {
	Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter) (string, error)
	List(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error)
}
