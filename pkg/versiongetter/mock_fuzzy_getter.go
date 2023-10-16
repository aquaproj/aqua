package versiongetter

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type MockFuzzyGetter struct {
	versions map[string]string
}

func NewMockFuzzyGetter(versions map[string]string) *MockFuzzyGetter {
	return &MockFuzzyGetter{
		versions: versions,
	}
}

func (g *MockFuzzyGetter) Get(ctx context.Context, _ *logrus.Entry, pkg *registry.PackageInfo, currentVersion string, useFinder bool) string {
	return g.versions[pkg.GetName()]
}
