package versiongetter

import (
	"context"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
)

type FuzzyGetter struct {
	fuzzyFinder FuzzyFinder
	gen         *Generator
}

func NewFuzzy(finder FuzzyFinder, gen *Generator) *FuzzyGetter {
	return &FuzzyGetter{
		fuzzyFinder: finder,
		gen:         gen,
	}
}

type MockFuzzyGetter struct {
	s string
}

func NewMockFuzzyGetter(s string) *MockFuzzyGetter {
	return &MockFuzzyGetter{
		s: s,
	}
}

func (g *MockFuzzyGetter) Get(ctx context.Context, _ *logrus.Entry, pkg *registry.PackageInfo, currentVersion string, useFinder bool) string {
	return g.s
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func (g *FuzzyGetter) Get(ctx context.Context, _ *logrus.Entry, pkg *registry.PackageInfo, currentVersion string, useFinder bool) string { //nolint:cyclop
	filters, err := createFilters(pkg)
	if err != nil {
		return ""
	}

	versionGetter := g.gen.Get(pkg)
	if versionGetter == nil {
		return ""
	}

	if useFinder { //nolint:nestif
		versions, err := versionGetter.List(ctx, pkg, filters)
		if err != nil {
			return ""
		}
		currentVersionIndex := 0
		if currentVersion != "" {
			for i, version := range versions {
				if version.Item == currentVersion {
					version.Item += " (*)"
					currentVersionIndex = i
					break
				}
			}
		}
		idx, err := g.fuzzyFinder.Find(versions, true)
		if err != nil {
			return ""
		}
		if idx == currentVersionIndex {
			return strings.TrimSuffix(versions[idx].Item, " (*)")
		}
		return versions[idx].Item
	}

	version, err := versionGetter.Get(ctx, pkg, filters)
	if err != nil {
		return ""
	}
	return version
}
