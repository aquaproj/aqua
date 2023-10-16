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

func TestGitHubReleaseVersionGetter_Get(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name     string
		releases map[string][]*github.RepositoryRelease
		pkg      *registry.PackageInfo
		filters  []*versiongetter.Filter
		isErr    bool
		version  string
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			releases: map[string][]*github.RepositoryRelease{
				"suzuki-shunsuke/tfcmt": {
					{
						TagName: ptr.String("v3.0.0"),
					},
					{
						TagName: ptr.String("v2.0.0"),
					},
					{
						TagName: ptr.String("v1.0.0"),
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
			ghReleaseClient := versiongetter.NewMockGitHubReleaseClient(d.releases)
			ghReleaseGetter := versiongetter.NewGitHubRelease(ghReleaseClient)
			version, err := ghReleaseGetter.Get(ctx, d.pkg, d.filters)
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

func TestGitHubReleaseVersionGetter_List(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		releases map[string][]*github.RepositoryRelease
		pkg      *registry.PackageInfo
		filters  []*versiongetter.Filter
		isErr    bool
		items    []*fuzzyfinder.Item
	}{
		{
			name: "normal",
			filters: []*versiongetter.Filter{
				{},
			},
			releases: map[string][]*github.RepositoryRelease{
				"suzuki-shunsuke/tfcmt": {
					{
						TagName: ptr.String("v3.0.0"),
						Body:    ptr.String("body(v3)"),
						HTMLURL: ptr.String("https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v3.0.0"),
					},
					{
						TagName: ptr.String("v2.0.0"),
						Body:    ptr.String("body(v2)"),
						HTMLURL: ptr.String("https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v2.0.0"),
					},
					{
						TagName: ptr.String("v1.0.0"),
						Body:    ptr.String("body(v1)"),
						HTMLURL: ptr.String("https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v1.0.0"),
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
					Preview: `v3.0.0

https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v3.0.0
body(v3)`,
				},
				{
					Item: "v2.0.0",
					Preview: `v2.0.0

https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v2.0.0
body(v2)`,
				},
				{
					Item: "v1.0.0",
					Preview: `v1.0.0

https://github.com/suzuki-shunsuke/tfcmt/releases/tag/v1.0.0
body(v1)`,
				},
			},
		},
	}

	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ghReleaseClient := versiongetter.NewMockGitHubReleaseClient(d.releases)
			ghReleaseGetter := versiongetter.NewGitHubRelease(ghReleaseClient)
			items, err := ghReleaseGetter.List(ctx, d.pkg, d.filters)
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
