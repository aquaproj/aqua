package versiongetter

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge/gitea"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge/gitlab"
)

// fakeForge answers from pages written in the test, and records what it was asked for.
type fakeForge struct {
	pages   [][]*forge.Release
	limit   int
	calls   int
	perPage int
	err     error
}

// MaxPerPage is the page size the getter asks this client for; zero is gitea's, which is
// what the tests that don't care about it read.
func (f *fakeForge) MaxPerPage() int {
	if f.perPage == 0 {
		return gitea.MaxPerPage
	}
	return f.perPage
}

func (f *fakeForge) ListReleases(_ context.Context, _, _, _ string, page, limit int) ([]*forge.Release, error) {
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

// fakeGitLab answers the way GitLab's client does, by project path, which is what tells
// the two clients apart.
type fakeGitLab struct {
	pages   [][]*forge.Release
	project string
	err     error
}

func (f *fakeGitLab) MaxPerPage() int {
	return gitlab.MaxPerPage
}

func (f *fakeGitLab) ListReleases(_ context.Context, _, project string, page, _ int) ([]*forge.Release, error) {
	f.project = project
	if f.err != nil {
		return nil, f.err
	}
	if page < 1 || page > len(f.pages) {
		return nil, nil
	}
	return f.pages[page-1], nil
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

func TestForgeReleaseVersionGetter_Get(t *testing.T) { //nolint:funlen
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	data := []struct {
		title string
		pages [][]*forge.Release
		pkg   *registry.PackageInfo
		exp   string
		isErr bool
	}{
		{
			title: "the newest release",
			pages: [][]*forge.Release{{
				{TagName: "v0.19.1"},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			title: "a prerelease is not a version to install",
			pages: [][]*forge.Release{{
				{TagName: "v0.21.0", Prerelease: true},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			title: "a draft is not a version to install",
			pages: [][]*forge.Release{{
				{TagName: "v0.21.0", Draft: true},
				{TagName: "v0.20.0"},
			}},
			exp: "v0.20.0",
		},
		{
			// A short page is not the last one: an instance can be configured to
			// answer with fewer items than were asked for.
			title: "the next page is read when the first holds nothing to install",
			pages: [][]*forge.Release{
				{{TagName: "v0.2.0", Draft: true}},
				{{TagName: "v0.1.0"}},
			},
			exp: "v0.1.0",
		},
		{
			title: "no release at all",
			pages: [][]*forge.Release{{}},
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
			getter := NewForgeRelease(&fakeForge{pages: d.pages}, nil)
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

// An instance answering the same page to every request is read no further than the cap,
// rather than forever.
func TestForgeReleaseVersionGetter_Get_samePageForever(t *testing.T) {
	t.Parallel()
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	client := &sameForgePage{release: &forge.Release{TagName: "v0.1.0", Draft: true}}
	getter := NewForgeRelease(client, nil)
	version, err := getter.Get(context.Background(), slog.New(slog.DiscardHandler), testPkg(), filters)
	if err != nil {
		t.Fatal(err)
	}
	if version != "" {
		t.Fatalf("wanted no version, got %s", version)
	}
	if client.calls != forge.MaxPages {
		t.Fatalf("wanted %d pages read, got %d", forge.MaxPages, client.calls)
	}
}

// sameForgePage answers every request with the same release, whatever page is asked for.
type sameForgePage struct {
	release *forge.Release
	calls   int
}

// MaxPerPage is gitea's, which is what the page bound is read against.
func (f *sameForgePage) MaxPerPage() int {
	return gitea.MaxPerPage
}

func (f *sameForgePage) ListReleases(_ context.Context, _, _, _ string, _, _ int) ([]*forge.Release, error) {
	f.calls++
	return []*forge.Release{f.release}, nil
}

func TestForgeReleaseVersionGetter_Get_error(t *testing.T) {
	t.Parallel()
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	getter := NewForgeRelease(&fakeForge{err: errors.New("the instance said no")}, nil)
	if _, err := getter.Get(context.Background(), slog.New(slog.DiscardHandler), testPkg(), filters); err == nil {
		t.Fatal("an error must be returned")
	}
}

func TestForgeReleaseVersionGetter_List(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}

	// A release listed twice, which a page boundary can do, is listed once.
	client := &fakeForge{pages: [][]*forge.Release{{
		{TagName: "v0.20.0"},
		{TagName: "v0.19.1"},
		{TagName: "v0.20.0"},
	}}}
	getter := NewForgeRelease(client, nil)
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
	client = &fakeForge{pages: [][]*forge.Release{{
		{TagName: "v0.20.0"},
		{TagName: "v0.19.1"},
	}}}
	getter = NewForgeRelease(client, nil)
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

func TestForgeItemNumPerPage(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		limit     int
		filterNum int
		exp       int
	}{
		{
			title: "no limit, so as much as the instance will answer with",
			exp:   gitea.MaxPerPage,
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
			exp:       gitea.MaxPerPage,
		},
		{
			title: "a limit larger than a page",
			limit: gitea.MaxPerPage + 1,
			exp:   gitea.MaxPerPage,
		},
	}
	// Each API caps a page at its own number, which is the one the getter asks for.
	if n := forgeItemNumPerPage(0, 0, gitlab.MaxPerPage); n != gitlab.MaxPerPage {
		t.Fatalf("wanted GitLab's page, got %d", n)
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if n := forgeItemNumPerPage(d.limit, d.filterNum, gitea.MaxPerPage); n != d.exp {
				t.Fatalf("wanted %d, got %d", d.exp, n)
			}
		})
	}
}

// Every type on a forge instance is read by the same getter, which is what picks the
// client that speaks the instance's API.
func TestGeneralVersionGetter_get_forge(t *testing.T) {
	t.Parallel()
	forgeGetter := NewForgeRelease(&fakeForge{}, &fakeGitLab{})
	getter := NewGeneralVersionGetter(nil, &GitHubTagVersionGetter{}, nil, forgeGetter, nil)
	for _, typ := range []string{registry.PkgInfoTypeForgejoRelease, registry.PkgInfoTypeGiteaRelease, registry.PkgInfoTypeGitLabRelease} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			if g := getter.get(&registry.PackageInfo{Type: typ, Host: "an.example.com"}); g != VersionGetter(forgeGetter) {
				t.Fatalf("wanted the forge getter for %s, got %T", typ, g)
			}
		})
	}
}

// Which client answers is the package's type, because the two APIs are not each other's.
func TestForgeReleaseVersionGetter_client(t *testing.T) {
	t.Parallel()
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.DiscardHandler)
	giteaClient := &fakeForge{pages: [][]*forge.Release{{{TagName: "v0.20.0"}}}}
	gitlabClient := &fakeGitLab{pages: [][]*forge.Release{{{TagName: "v1.122.0"}}}}
	getter := NewForgeRelease(giteaClient, gitlabClient)

	for _, d := range []struct {
		typ string
		exp string
	}{
		{typ: registry.PkgInfoTypeForgejoRelease, exp: "v0.20.0"},
		{typ: registry.PkgInfoTypeGiteaRelease, exp: "v0.20.0"},
		{typ: registry.PkgInfoTypeGitLabRelease, exp: "v1.122.0"},
	} {
		t.Run(d.typ, func(t *testing.T) {
			t.Parallel()
			version, err := getter.Get(context.Background(), logger,
				&registry.PackageInfo{Type: d.typ, Host: "an.example.com", RepoOwner: "an-owner", RepoName: "a-repo"}, filters)
			if err != nil {
				t.Fatal(err)
			}
			if version != d.exp {
				t.Fatalf("wanted %s, which is what that forge's client answers, got %s", d.exp, version)
			}
			if d.typ == registry.PkgInfoTypeGitLabRelease && gitlabClient.project != "an-owner/a-repo" {
				t.Fatalf("wanted the project GitLab reads as one id, got %q", gitlabClient.project)
			}
		})
	}
}

// A getter built without the client a type is read by refuses that type rather than
// calling nothing.
func TestForgeReleaseVersionGetter_client_missing(t *testing.T) {
	t.Parallel()
	filters, err := createFilters(testPkg())
	if err != nil {
		t.Fatal(err)
	}
	getter := NewForgeRelease(&fakeForge{}, nil)
	if _, err := getter.Get(context.Background(), slog.New(slog.DiscardHandler),
		&registry.PackageInfo{Type: registry.PkgInfoTypeGitLabRelease, Host: "gitlab.com"}, filters); err == nil {
		t.Fatal("an error must be returned")
	}
}

// Without one, nothing is returned rather than a getter that can't be called.
func TestGeneralVersionGetter_get_noForge(t *testing.T) {
	t.Parallel()
	getter := NewGeneralVersionGetter(nil, &GitHubTagVersionGetter{}, nil, nil, nil)
	if g := getter.get(&registry.PackageInfo{Type: registry.PkgInfoTypeGiteaRelease, Host: "an.example.com"}); g != nil {
		t.Fatalf("wanted no getter, got %T", g)
	}
}
