package versiongetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

type MockFuzzyGetter struct {
	versions map[string]string
}

func NewMockFuzzyGetter(versions map[string]string) *MockFuzzyGetter {
	return &MockFuzzyGetter{
		versions: versions,
	}
}

func (g *MockFuzzyGetter) Get(_ context.Context, _ *slog.Logger, pkg *registry.PackageInfo, _ string, _ bool, _ int) string {
	return g.versions[pkg.GetName()]
}
