package g2_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/google/go-cmp/cmp"
)

type fakeTreeGetter struct {
	sha  string
	tree *github.Tree
	err  error
}

func (g *fakeTreeGetter) GetTree(_ context.Context, _, _, sha string, _ bool) (*github.Tree, *github.Response, error) {
	g.sha = sha
	return g.tree, nil, g.err
}

func tree(truncated bool, entries ...*github.TreeEntry) *github.Tree {
	return &github.Tree{Truncated: &truncated, Entries: entries}
}

func entry(typ, path string) *github.TreeEntry {
	return &github.TreeEntry{Type: &typ, Path: &path}
}

// version is what a version looks like in the tree: a directory holding the generated
// file.
func version(v string) []*github.TreeEntry {
	return []*github.TreeEntry{entry("tree", v), entry("blob", v+"/registry-1.json")}
}

func entries(groups ...[]*github.TreeEntry) []*github.TreeEntry {
	var out []*github.TreeEntry
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

func TestVersionLister_List(t *testing.T) {
	t.Parallel()
	gh := &fakeTreeGetter{
		tree: tree(false, entries(
			version("v1.0.0"),
			version("v10.0.0"),
			version("v2.0.0"),
			// The branch carries a README and workflows too, which aren't versions.
			[]*github.TreeEntry{entry("blob", "README.md")},
		)...),
	}
	lister := g2.NewVersionLister(gh, "", "")

	got, err := lister.List(t.Context(), slog.New(slog.DiscardHandler), "cli/cli")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"v1.0.0", "v10.0.0", "v2.0.0"}, got); diff != "" {
		t.Errorf("the versions are wrong (-want +got):\n%s", diff)
	}
	// The package is the branch and the versions are a directory in it.
	if diff := cmp.Diff("pkg_cli_2fcli:versions", gh.sha); diff != "" {
		t.Errorf("the tree is wrong (-want +got):\n%s", diff)
	}
}

// A partial list would silently drop versions, and the ones dropped needn't be the
// oldest, because the tree is ordered by name rather than by version.
func TestVersionLister_List_truncated(t *testing.T) {
	t.Parallel()
	lister := g2.NewVersionLister(&fakeTreeGetter{tree: tree(true, version("v1.0.0")...)}, "", "")
	if _, err := lister.List(t.Context(), slog.New(slog.DiscardHandler), "cli/cli"); err == nil {
		t.Fatal("an error must be returned")
	}
}

func TestVersionLister_List_error(t *testing.T) {
	t.Parallel()
	lister := g2.NewVersionLister(&fakeTreeGetter{err: errors.New("404")}, "", "")
	if _, err := lister.List(t.Context(), slog.New(slog.DiscardHandler), "cli/cli"); err == nil {
		t.Fatal("an error must be returned")
	}
}

// A version whose tag has a slash in it is escaped into one directory, the way a
// package name is escaped into one branch.
func TestVersionLister_List_escaped(t *testing.T) {
	t.Parallel()
	gh := &fakeTreeGetter{
		tree: tree(false, entries(
			version("kustomize_2fv5.8.1"),
			version("kustomize_2fv5.8.2"),
		)...),
	}
	lister := g2.NewVersionLister(gh, "", "")

	got, err := lister.List(t.Context(), slog.New(slog.DiscardHandler), "kubernetes-sigs/kustomize")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"kustomize/v5.8.1", "kustomize/v5.8.2"}, got); diff != "" {
		t.Errorf("the versions are wrong (-want +got):\n%s", diff)
	}
}

// The layout once wrote those slashes as directories, which made every kustomize
// release a directory called "kustomize". Offering that as a version wrote it into
// aqua.yaml, where nothing could install it: no release is tagged "kustomize".
func TestVersionLister_List_nested(t *testing.T) {
	t.Parallel()
	gh := &fakeTreeGetter{
		tree: tree(false,
			entry("tree", "kustomize"),
			entry("tree", "kustomize/v5.8.1"),
			entry("blob", "kustomize/v5.8.1/registry-1.json"),
		),
	}
	lister := g2.NewVersionLister(gh, "", "")

	got, err := lister.List(t.Context(), slog.New(slog.DiscardHandler), "kubernetes-sigs/kustomize")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{}, got); diff != "" {
		t.Errorf("nothing there is a version (-want +got):\n%s", diff)
	}
}
