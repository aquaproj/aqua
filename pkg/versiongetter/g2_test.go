package versiongetter_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/google/go-cmp/cmp"
	gogithub "github.com/google/go-github/v92/github"
)

type fakeTreeGetter struct {
	versions []string
	// notFound makes it answer the way GitHub does for a package aqua-registry-g2
	// has no branch for.
	notFound bool
}

func (g *fakeTreeGetter) GetTree(_ context.Context, _, _, _ string, _ bool) (*github.Tree, *github.Response, error) {
	if g.notFound {
		return nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
			&gogithub.ErrorResponse{Message: "Not Found"}
	}
	truncated := false
	tree := &github.Tree{Truncated: &truncated}
	for _, v := range g.versions {
		typ := "tree"
		tree.Entries = append(tree.Entries, &github.TreeEntry{Type: &typ, Path: &v})
	}
	return tree, nil, nil
}

func newG2(versions ...string) *versiongetter.G2VersionGetter {
	return versiongetter.NewG2(g2.NewVersionLister(&fakeTreeGetter{versions: versions}, "", ""))
}

// newG2NotFound builds a getter for a package aqua-registry-g2 has no branch for.
func newG2NotFound() *versiongetter.G2VersionGetter {
	return versiongetter.NewG2(g2.NewVersionLister(&fakeTreeGetter{notFound: true}, "", ""))
}

func items(versions ...string) []*fuzzyfinder.Item {
	return fuzzyfinder.ConvertStringsToItems(versions)
}

// The branch orders versions by name, so the getter has to put them in version order
// itself: "v10.0.0" comes after "v2.0.0" by name and before it by version.
func TestG2VersionGetter_List(t *testing.T) {
	t.Parallel()
	pkg := &registry.PackageInfo{Name: "cli/cli"}
	logger := slog.New(slog.DiscardHandler)

	got, err := newG2("v1.0.0", "v10.0.0", "v2.0.0", "v2.1.0-rc.1").List(t.Context(), logger, pkg, nil, -1)
	if err != nil {
		t.Fatal(err)
	}
	// A prerelease sorts below every released version, the same as on the registry path.
	if diff := cmp.Diff(items("v10.0.0", "v2.0.0", "v1.0.0", "v2.1.0-rc.1"), got); diff != "" {
		t.Errorf("the versions are wrong (-want +got):\n%s", diff)
	}
}

func TestG2VersionGetter_List_limit(t *testing.T) {
	t.Parallel()
	got, err := newG2("v1.0.0", "v10.0.0", "v2.0.0").List(t.Context(), slog.New(slog.DiscardHandler), &registry.PackageInfo{Name: "cli/cli"}, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(items("v10.0.0", "v2.0.0"), got); diff != "" {
		t.Errorf("the versions are wrong (-want +got):\n%s", diff)
	}
}

func TestG2VersionGetter_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		versions []string
		want     string
	}{
		{
			name:     "the newest version",
			versions: []string{"v1.0.0", "v10.0.0", "v2.0.0"},
			want:     "v10.0.0",
		},
		{
			// A prerelease is not what "latest" means, so a released version wins
			// however new the prerelease is.
			name:     "a prerelease isn't the latest",
			versions: []string{"v1.0.0", "v2.0.0-rc.1"},
			want:     "v1.0.0",
		},
		{
			name:     "no version at all",
			versions: nil,
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := newG2(tt.versions...).Get(t.Context(), slog.New(slog.DiscardHandler), &registry.PackageInfo{Name: "cli/cli"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
