package versiongetter_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
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
						Name: ptr.String("v3.0.0"),
					},
					{
						Name: ptr.String("v2.0.0"),
					},
					{
						Name: ptr.String("v1.0.0"),
					},
				},
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			version: "v3.0.0",
		},
	}

	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ghTagClient := versiongetter.NewMockGitHubTagClient(d.tags)
			ghTagGetter := versiongetter.NewGitHubTag(ghTagClient)
			version, err := ghTagGetter.Get(ctx, d.pkg, d.filters)
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
						Name: ptr.String("v3.0.0"),
					},
					{
						Name: ptr.String("v2.0.0"),
					},
					{
						Name: ptr.String("v1.0.0"),
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

	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ghTagClient := versiongetter.NewMockGitHubTagClient(d.tags)
			ghTagGetter := versiongetter.NewGitHubTag(ghTagClient)
			items, err := ghTagGetter.List(ctx, d.pkg, d.filters)
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
				t.Fatalf(diff)
			}
		})
	}
}
