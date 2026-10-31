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

// fakeForgeTags answers from pages written in the test, and records what it was asked for.
type fakeForgeTags struct {
	pages [][]*forge.Tag
	limit int
	calls int
	err   error
}

func (f *fakeForgeTags) MaxPerPage() int {
	return gitea.MaxPerPage
}

func (f *fakeForgeTags) ListTags(_ context.Context, _, _, _ string, page, limit int) ([]*forge.Tag, error) {
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

// fakeGitLabTags answers the way GitLab's client does, by project path.
type fakeGitLabTags struct {
	pages   [][]*forge.Tag
	project string
}

func (f *fakeGitLabTags) MaxPerPage() int {
	return gitlab.MaxPerPage
}

func (f *fakeGitLabTags) ListTags(_ context.Context, _, project string, page, _ int) ([]*forge.Tag, error) {
	f.project = project
	if page < 1 || page > len(f.pages) {
		return nil, nil
	}
	return f.pages[page-1], nil
}

// testTagPkg is Zig: an http package whose versions are Codeberg's tags.
func testTagPkg() *registry.PackageInfo {
	return &registry.PackageInfo{
		Type:          registry.PkgInfoTypeHTTP,
		Host:          "codeberg.org",
		RepoOwner:     "ziglang",
		RepoName:      "zig",
		VersionSource: registry.VersionSourceForgejoTag,
		URL:           "https://ziglang.org/download/{{.Version}}/zig-{{.Arch}}-{{.OS}}-{{.Version}}.{{.Format}}",
	}
}

func tags(names ...string) []*forge.Tag {
	out := make([]*forge.Tag, 0, len(names))
	for _, name := range names {
		out = append(out, &forge.Tag{Name: name})
	}
	return out
}

func TestForgeTagVersionGetter_Get(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	data := []struct {
		title  string
		filter string
		pages  [][]*forge.Tag
		exp    string
	}{
		{
			title: "the newest version, not the first tag",
			pages: [][]*forge.Tag{tags("0.16.0", "0.17.0", "0.15.2")},
			exp:   "0.17.0",
		},
		{
			title: "a prerelease is left out",
			pages: [][]*forge.Tag{tags("0.18.0-dev.1", "0.17.0")},
			exp:   "0.17.0",
		},
		{
			title:  "a page with nothing allowed is followed by the next",
			filter: `not (Version startsWith "nightly")`,
			pages:  [][]*forge.Tag{tags("nightly-2026-10-22"), tags("0.17.0")},
			exp:    "0.17.0",
		},
		{
			title: "no tag",
			exp:   "",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkg := testTagPkg()
			pkg.VersionFilter = d.filter
			filters, err := createFilters(pkg)
			if err != nil {
				t.Fatal(err)
			}
			getter := NewForgeTag(&fakeForgeTags{pages: d.pages}, nil)
			version, err := getter.Get(context.Background(), logger, pkg, filters)
			if err != nil {
				t.Fatal(err)
			}
			if version != d.exp {
				t.Fatalf("wanted %q, got %q", d.exp, version)
			}
		})
	}
}

func TestForgeTagVersionGetter_Get_error(t *testing.T) {
	t.Parallel()
	pkg := testTagPkg()
	filters, err := createFilters(pkg)
	if err != nil {
		t.Fatal(err)
	}
	getter := NewForgeTag(&fakeForgeTags{err: errors.New("the instance said no")}, nil)
	if _, err := getter.Get(context.Background(), slog.New(slog.DiscardHandler), pkg, filters); err == nil {
		t.Fatal("an error must be returned")
	}
}

func TestForgeTagVersionGetter_List(t *testing.T) {
	t.Parallel()
	pkg := testTagPkg()
	filters, err := createFilters(pkg)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.DiscardHandler)
	client := &fakeForgeTags{pages: [][]*forge.Tag{tags("0.17.0", "0.16.0"), tags("0.16.0", "0.15.2")}}
	getter := NewForgeTag(client, nil)
	items, err := getter.List(context.Background(), logger, pkg, filters, 0)
	if err != nil {
		t.Fatal(err)
	}
	// A tag on two pages is listed once.
	if len(items) != 3 || items[0].Item != "0.17.0" || items[2].Item != "0.15.2" {
		t.Fatalf("wanted every tag once, in the instance's order, got %+v", items)
	}

	client = &fakeForgeTags{pages: [][]*forge.Tag{tags("0.17.0", "0.16.0", "0.15.2")}}
	getter = NewForgeTag(client, nil)
	items, err = getter.List(context.Background(), logger, pkg, filters, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("wanted the limit, got %d", len(items))
	}
	// A filter can leave a tag out, so a whole page is asked for rather than the limit.
	if client.limit != gitea.MaxPerPage {
		t.Fatalf("wanted a whole page, got %d", client.limit)
	}
}

// Which client answers is the version source, because an http package is on no forge.
func TestForgeTagVersionGetter_client(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	for _, d := range []struct {
		src string
		exp string
	}{
		{src: registry.VersionSourceForgejoTag, exp: "0.17.0"},
		{src: registry.VersionSourceGiteaTag, exp: "0.17.0"},
		{src: registry.VersionSourceGitLabTag, exp: "v1.10.0"},
	} {
		t.Run(d.src, func(t *testing.T) {
			t.Parallel()
			gitlabClient := &fakeGitLabTags{pages: [][]*forge.Tag{tags("v1.9.0", "v1.10.0")}}
			getter := NewForgeTag(&fakeForgeTags{pages: [][]*forge.Tag{tags("0.17.0")}}, gitlabClient)
			pkg := testTagPkg()
			pkg.VersionSource = d.src
			filters, err := createFilters(pkg)
			if err != nil {
				t.Fatal(err)
			}
			version, err := getter.Get(context.Background(), logger, pkg, filters)
			if err != nil {
				t.Fatal(err)
			}
			if version != d.exp {
				t.Fatalf("wanted %s, which is what that forge's client answers, got %s", d.exp, version)
			}
			if d.src == registry.VersionSourceGitLabTag && gitlabClient.project != "ziglang/zig" {
				t.Fatalf("wanted the project GitLab reads as one id, got %q", gitlabClient.project)
			}
		})
	}
}

func TestForgeTagVersionGetter_client_missing(t *testing.T) {
	t.Parallel()
	pkg := testTagPkg()
	pkg.VersionSource = registry.VersionSourceGitLabTag
	if _, err := NewForgeTag(&fakeForgeTags{}, nil).Get(context.Background(), slog.New(slog.DiscardHandler), pkg, nil); err == nil {
		t.Fatal("an error must be returned")
	}
	pkg = testTagPkg()
	pkg.Host = ""
	if _, err := NewForgeTag(&fakeForgeTags{}, nil).Get(context.Background(), slog.New(slog.DiscardHandler), pkg, nil); err == nil {
		t.Fatal("an error must be returned for a package that names no instance")
	}
}

// A forge's tags are read whatever the type is: before the type's own getter, and instead
// of GitHub's.
func TestGeneralVersionGetter_get_forgeTag(t *testing.T) {
	t.Parallel()
	forgeRelease := NewForgeRelease(&fakeForge{}, &fakeGitLab{})
	forgeTag := NewForgeTag(&fakeForgeTags{}, &fakeGitLabTags{})
	getter := NewGeneralVersionGetter(nil, &GitHubTagVersionGetter{}, nil, forgeRelease, forgeTag, nil)

	if g := getter.get(testTagPkg()); g != VersionGetter(forgeTag) {
		t.Fatalf("wanted the forge tag getter for an http package, got %T", g)
	}
	pkg := testPkg()
	pkg.VersionSource = registry.VersionSourceForgejoTag
	if g := getter.get(pkg); g != VersionGetter(forgeTag) {
		t.Fatalf("wanted the forge tag getter for a release saying so, got %T", g)
	}
	if g := getter.get(testPkg()); g != VersionGetter(forgeRelease) {
		t.Fatalf("wanted the forge release getter for a release saying nothing, got %T", g)
	}
	pkg = testTagPkg()
	pkg.RepoName = ""
	if g := getter.get(pkg); g != nil {
		t.Fatalf("wanted no getter for a package naming no repository, got %T", g)
	}
	if g := NewGeneralVersionGetter(nil, &GitHubTagVersionGetter{}, nil, nil, nil, nil).get(testTagPkg()); g != nil {
		t.Fatalf("wanted no getter without one, got %T", g)
	}
}
