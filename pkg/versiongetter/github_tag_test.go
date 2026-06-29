package versiongetter_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/expr"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
)

func TestGitHubTagVersionGetter_Get(t *testing.T) { //nolint:dupl
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
			// Reproduces a behavioural asymmetry between github_release and
			// github_tag version sources.
			//
			// github_release's filterRelease() hard-rejects any release whose
			// `prerelease` flag is true (pkg/versiongetter/github_release.go),
			// so pre-release tags never enter the candidate list. The Get()
			// loop then pages on through ListReleases until a stable release
			// is found.
			//
			// github_tag's filterTag() has no equivalent check. When the most
			// recent page of tags contains only pre-releases (e.g. >PerPage
			// pre-release tags accumulated since the last stable), candidates
			// fills with pre-releases, len(candidates) > 0 short-circuits the
			// Get() loop, and the latest stable on a later page is never
			// reached. compareRelease()'s stable-vs-pre tiebreaker doesn't
			// rescue this because both sides of the comparison are pre-release.
			//
			// Expected (matching github_release behaviour): "v1.0.0".
			// Observed (current github_tag behaviour): "v1.0.1-beta.30".
			name: "prereleases dominate page 1, stable on page 2 — should return stable",
			filters: []*versiongetter.Filter{
				{},
			},
			tags: map[string][]*github.RepositoryTag{
				"kunobi-ninja/kunobi": tagListWithPrereleaseDominantFirstPage(),
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "kunobi-ninja",
				RepoName:  "kunobi",
			},
			version: "v1.0.0",
		},
		{
			// Same data as the previous case, but with a version_filter that
			// excludes the rc/alpha/beta pre-release pattern. filterTag now
			// rejects all of page 1's candidates, len(candidates) stays 0,
			// the Get() loop pages on to page 2, and the stable is found.
			//
			// This documents the workaround that registry authors currently
			// have to apply manually (e.g. apache/maven, the package whose
			// version_filter expression matches this one), and that the
			// asymmetry above means is mandatory for github_tag packages but
			// unnecessary for github_release packages.
			name: "version_filter excluding pre-releases recovers the stable",
			filters: func() []*versiongetter.Filter {
				f, err := expr.CompileVersionFilter(`not (Version matches "-(rc|alpha|beta)")`)
				if err != nil {
					panic(err)
				}
				return []*versiongetter.Filter{{Filter: f}}
			}(),
			tags: map[string][]*github.RepositoryTag{
				"kunobi-ninja/kunobi": tagListWithPrereleaseDominantFirstPage(),
			},
			pkg: &registry.PackageInfo{
				RepoOwner: "kunobi-ninja",
				RepoName:  "kunobi",
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

// tagListWithPrereleaseDominantFirstPage returns a tag list that triggers
// the github_tag prerelease-pagination edge documented above:
// 30 pre-release tags (filling the entire first page given the hardcoded
// PerPage = 30 in github_tag.go) followed by 1 stable tag on page 2.
// Tags are ordered most-recent-first, matching real GitHub API output.
func tagListWithPrereleaseDominantFirstPage() []*github.RepositoryTag {
	tags := make([]*github.RepositoryTag, 0, 31)
	// Page 1 — 30 pre-releases.
	for i := 30; i >= 1; i-- {
		tags = append(tags, &github.RepositoryTag{
			Name: new(fmt.Sprintf("v1.0.1-beta.%d", i)),
		})
	}
	// Page 2 — the stable release the user actually wants.
	tags = append(tags, &github.RepositoryTag{
		Name: new("v1.0.0"),
	})
	return tags
}
