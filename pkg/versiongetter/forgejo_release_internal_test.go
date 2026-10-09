package versiongetter

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forgejo"
)

// fakeForgejo answers from pages written in the test, and records what it was asked for.
type fakeForgejo struct {
	pages [][]*forgejo.Release
	limit int
	calls int
	err   error
}

func (f *fakeForgejo) ListReleases(_ context.Context, _, _, _ string, page, limit int) ([]*forgejo.Release, error) {
	f.calls++
	f.limit = limit
	if f.err != nil {
		return nil, f.err
	}
	if page < 1 || page > len(f.pages) {
		return nil, nil
	}
	return f.pages[page-1], nil
}

func fullPage(tags ...string) []*forgejo.Release {
	releases := make([]*forgejo.Release, 0, forgejo.MaxPerPage)
	for _, tag := range tags {
		releases = append(releases, &forgejo.Release{TagName: tag})
	}
	// A page shorter than the limit is the last one, so a page that is meant to be
	// followed by another has to be full.
	for len(releases) < forgejo.MaxPerPage {
		releases = append(releases, &forgejo.Release{TagName: "v0.0.0", Draft: true})
	}
	return releases
}

func testPkg() *registry.PackageInfo {
	return &registry.PackageInfo{
		Type:      registry.PkgInfoTypeForgejoRelease,
		Host:      "codeberg.org",
		RepoOwner: "mergiraf",
		RepoName:  "mergiraf",
		Asset:     "mergiraf.{{.Format}}",
	}
}

func TestForgejoReleaseVersionGetter_Get(t *testing.T) { //nolint:funlen
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	data := []struct {
		title string
		pages [][]*forgejo.Release
		pkg   *registry.PackageInfo
		exp   string
		isErr bool
	}{
		{
			title: "the newest release",
			pages: [][]*forgejo.Release{{
				{TagName: "v0.19.1"},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			title: "a prerelease is not a version to install",
			pages: [][]*forgejo.Release{{
				{TagName: "v0.21.0", Prerelease: true},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			title: "a draft is not a version to install",
			pages: [][]*forgejo.Release{{
				{TagName: "v0.21.0", Draft: true},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			title: "the next page is read when the first holds nothing to install",
			pages: [][]*forgejo.Release{
				fullPage(),
				{{TagName: "v0.1.0"}},
			},
			exp: "v0.1.0",
		},
		{
			title: "no release at all",
			pages: [][]*forgejo.Release{{}},
			exp:   "",
		},
		{
			title: "the instance is required",
			pkg: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkg := d.pkg
			if pkg == nil {
				pkg = testPkg()
			}
			getter := NewForgejoRelease(&fakeForgejo{pages: d.pages})
			version, err := getter.Get(context.Background(), logger, pkg, filters)
			if d.isErr {
				if err == nil {
					t.Fatal("an error must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if version != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, version)
			}
		})
	}
}

func TestForgejoReleaseVersionGetter_Get_error(t *testing.T) {
	t.Parallel()
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	getter := NewForgejoRelease(&fakeForgejo{err: errors.New("the instance said no")})
	if _, err := getter.Get(context.Background(), slog.New(slog.DiscardHandler), testPkg(), filters); err == nil {
		t.Fatal("an error must be returned")
	}
}

func TestForgejoReleaseVersionGetter_List(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}

	// A release listed twice, which a page boundary can do, is listed once.
	client := &fakeForgejo{pages: [][]*forgejo.Release{{
		{TagName: "v0.20.0"},
		{TagName: "v0.19.1"},
		{TagName: "v0.20.0"},
	}}}
	getter := NewForgejoRelease(client)
	items, err := getter.List(context.Background(), logger, testPkg(), filters, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("wanted 2 versions, got %d", len(items))
	}
	if items[0].Item != "v0.20.0" || items[1].Item != "v0.19.1" {
		t.Fatalf("wanted the releases in the order the instance gave them, got %s and %s", items[0].Item, items[1].Item)
	}

	// The limit is the number of versions shown, not the number read.
	client = &fakeForgejo{pages: [][]*forgejo.Release{{
		{TagName: "v0.20.0"},
		{TagName: "v0.19.1"},
	}}}
	getter = NewForgejoRelease(client)
	items, err = getter.List(context.Background(), logger, testPkg(), filters, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("wanted 1 version, got %d", len(items))
	}
	if items[0].Item != "v0.20.0" {
		t.Fatalf("wanted the newest version, got %s", items[0].Item)
	}
}

func TestForgejoItemNumPerPage(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		limit     int
		filterNum int
		exp       int
	}{
		{
			title: "no limit, so as much as the instance will answer with",
			exp:   forgejo.MaxPerPage,
		},
		{
			title: "a limit smaller than a page, and nothing filtered out",
			limit: 10,
			exp:   10,
		},
		{
			title:     "a filter, so the page is full however small the limit is",
			limit:     10,
			filterNum: 1,
			exp:       forgejo.MaxPerPage,
		},
		{
			title: "a limit larger than a page",
			limit: forgejo.MaxPerPage + 1,
			exp:   forgejo.MaxPerPage,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if n := forgejoItemNumPerPage(d.limit, d.filterNum); n != d.exp {
				t.Fatalf("wanted %d, got %d", d.exp, n)
			}
		})
	}
}
