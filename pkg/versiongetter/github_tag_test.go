package versiongetter_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
)

func TestGitHubTagVersionGetter_Get(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		tags    map[string][]*github.RepositoryTag
		pkg     *registry.PackageInfo
		filters []*versiongetter.Filter
		isErr   bool
		version string
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			tags: map[string][]*github.RepositoryTag{
				"suzuki-shunsuke/tfcmt": {
					{
						Name: new("v3.0.0"),
					},
					{
						Name: new("v2.0.0"),
					},
					{
						Name: new("v1.0.0"),
					},
				},
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			version: "v3.0.0",
		},
		{
			name: "prereleases dominated page 1, stable on page 2",
			filters: []*versiongetter.Filter{
				{},
			},
			tags: map[string][]*github.RepositoryTag{
				"testuser/testpkg": {
					{
						Name: new("v1.0.1-beta.30"),
					},
					{
						Name: new("v1.0.1-beta.29"),
					},
					{
						Name: new("v1.0.1-beta.28"),
					},
					{
						Name: new("v1.0.1-beta.27"),
					},
					{
						Name: new("v1.0.1-beta.26"),
					},
					{
						Name: new("v1.0.1-beta.25"),
					},
					{
						Name: new("v1.0.1-beta.24"),
					},
					{
						Name: new("v1.0.1-beta.23"),
					},
					{
						Name: new("v1.0.1-beta.22"),
					},
					{
						Name: new("v1.0.1-beta.21"),
					},
					{
						Name: new("v1.0.1-beta.20"),
					},
					{
						Name: new("v1.0.1-beta.19"),
					},
					{
						Name: new("v1.0.1-beta.18"),
					},
					{
						Name: new("v1.0.1-beta.17"),
					},
					{
						Name: new("v1.0.1-beta.16"),
					},
					{
						Name: new("v1.0.1-beta.15"),
					},
					{
						Name: new("v1.0.1-beta.14"),
					},
					{
						Name: new("v1.0.1-beta.13"),
					},
					{
						Name: new("v1.0.1-beta.12"),
					},
					{
						Name: new("v1.0.1-beta.11"),
					},
					{
						Name: new("v1.0.1-beta.10"),
					},
					{
						Name: new("v1.0.1-beta.9"),
					},
					{
						Name: new("v1.0.1-beta.8"),
					},
					{
						Name: new("v1.0.1-beta.7"),
					},
					{
						Name: new("v1.0.1-beta.6"),
					},
					{
						Name: new("v1.0.1-beta.5"),
					},
					{
						Name: new("v1.0.1-beta.4"),
					},
					{
						Name: new("v1.0.1-beta.3"),
					},
					{
						Name: new("v1.0.1-beta.2"),
					},
					{
						Name: new("v1.0.1-beta.1"),
					},
					{
						Name: new("v1.0.0"),
					},
				},
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "testuser",
				RepoName:  "testpkg",
			},
			version: "v1.0.0",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			ghTagClient := versiongetter.NewMockGitHubTagClient(d.tags)
			ghTagGetter := versiongetter.NewGitHubTag(ghTagClient)
			version, err := ghTagGetter.Get(ctx, slog.New(slog.DiscardHandler), d.pkg, d.filters)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if version != d.version {
				t.Fatalf("wanted %s, got %s", d.version, version)
			}
		})
	}
}

func TestGitHubTagVersionGetter_List(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		tags    map[string][]*github.RepositoryTag
		pkg     *registry.PackageInfo
		filters []*versiongetter.Filter
		isErr   bool
		items   []*fuzzyfinder.Item
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			tags: map[string][]*github.RepositoryTag{
				"suzuki-shunsuke/tfcmt": {
					{
						Name: new("v3.0.0"),
					},
					{
						Name: new("v2.0.0"),
					},
					{
						Name: new("v1.0.0"),
					},
				},
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			items: []*fuzzyfinder.Item{
				{
					Item: "v3.0.0",
				},
				{
					Item: "v2.0.0",
				},
				{
					Item: "v1.0.0",
				},
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			ghTagClient := versiongetter.NewMockGitHubTagClient(d.tags)
			ghTagGetter := versiongetter.NewGitHubTag(ghTagClient)
			items, err := ghTagGetter.List(ctx, slog.New(slog.DiscardHandler), d.pkg, d.filters, -1)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(items, d.items); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
