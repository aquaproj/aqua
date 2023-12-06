package versiongetter

import (
	"context"
	"errors"
)

type MockCargoClient struct {
	versions map[string][]string
}

func NewMockCargoClient(versions map[string][]string) *MockCargoClient {
	return &MockCargoClient{
		versions: versions,
	}
}

func (g *MockCargoClient) ListVersions(ctx context.Context, crate string) ([]string, error) {
	versions, ok := g.versions[crate]
	if !ok {
		return nil, errors.New("crate isn't found")
	}
	return versions, nil
}

func (g *MockCargoClient) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	versions, ok := g.versions[crate]
	if !ok {
		return "", errors.New("crate isn't found")
	}
	return versions[0], nil
}
