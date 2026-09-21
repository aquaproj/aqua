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

func TestVersionLister_List(t *testing.T) {
	t.Parallel()
	gh := &fakeTreeGetter{
		tree: tree(false,
			entry("tree", "v1.0.0"),
			entry("tree", "v10.0.0"),
			entry("tree", "v2.0.0"),
			// The branch carries a README and workflows too, which aren't versions.
			entry("blob", "README.md"),
		),
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
	lister := g2.NewVersionLister(&fakeTreeGetter{tree: tree(true, entry("tree", "v1.0.0"))}, "", "")
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
