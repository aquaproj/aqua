package g2_test

import (
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

func TestNewIndexPackage(t *testing.T) {
	t.Parallel()
	got := g2.NewIndexPackage("cli/cli", &g2.Config{
		PackageInfo: &registry.PackageInfo{
			Description: "GitHub's official command line tool",
			Link:        "https://cli.github.com/",
			SearchWords: []string{"github"},
			Aliases:     []*registry.Alias{{Name: "gh"}, {Name: ""}},
		},
	})
	want := &g2.IndexPackage{
		Name:        "cli/cli",
		Description: "GitHub's official command line tool",
		Link:        "https://cli.github.com/",
		SearchWords: []string{"github"},
		// The empty alias is dropped: it names nothing to search for.
		Aliases: []string{"gh"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the entry is wrong (-want +got):\n%s", diff)
	}
}

// The package name comes from the branch, so an entry can be made for a package whose
// definition says nothing about itself.
func TestNewIndexPackage_noConfig(t *testing.T) {
	t.Parallel()
	got := g2.NewIndexPackage("foo/foo", nil)
	if diff := cmp.Diff(&g2.IndexPackage{Name: "foo/foo"}, got); diff != "" {
		t.Errorf("the entry is wrong (-want +got):\n%s", diff)
	}
}

// A package added later belongs where its name puts it, so that an update changes the
// lines it adds and no others.
func TestIndex_Add(t *testing.T) {
	t.Parallel()
	index := &g2.Index{Packages: []*g2.IndexPackage{
		{Name: "aquaproj/aqua"},
		{Name: "cli/cli"},
	}}
	index.Add(&g2.IndexPackage{Name: "bufbuild/buf"})

	names := make([]string, len(index.Packages))
	for i, pkg := range index.Packages {
		names[i] = pkg.Name
	}
	if diff := cmp.Diff([]string{"aquaproj/aqua", "bufbuild/buf", "cli/cli"}, names); diff != "" {
		t.Errorf("the order is wrong (-want +got):\n%s", diff)
	}
}

func TestIndex_roundTrip(t *testing.T) {
	t.Parallel()
	index := &g2.Index{Packages: []*g2.IndexPackage{
		{Name: "cli/cli", Description: "GitHub's official command line tool", Aliases: []string{"gh"}},
	}}
	s, err := index.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Error("the file must end with a newline")
	}
	got, err := g2.ReadIndex(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(index, got); diff != "" {
		t.Errorf("the index is wrong (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(map[string]struct{}{"cli/cli": {}}, got.Names()); diff != "" {
		t.Errorf("the names are wrong (-want +got):\n%s", diff)
	}
}

// A package leaves the catalogue when its repository is renamed: the entry under the new
// name carries the old one as an alias, and a name that is both a package and another
// package's alias is a state the catalogue is checked for.
func TestIndex_Remove(t *testing.T) {
	t.Parallel()
	index := &g2.Index{Packages: []*g2.IndexPackage{
		{Name: "cli/cli"},
		{Name: "sst/opencode"},
		nil,
		{Name: "anomalyco/opencode", Aliases: []string{"sst/opencode"}},
	}}
	index.Remove("sst/opencode")

	got := make([]string, 0, len(index.Packages))
	for _, pkg := range index.Packages {
		if pkg == nil {
			continue
		}
		got = append(got, pkg.Name)
	}
	if diff := cmp.Diff([]string{"cli/cli", "anomalyco/opencode"}, got); diff != "" {
		t.Errorf("the catalogue is wrong (-want +got):\n%s", diff)
	}
	// Removing what isn't there is not an error: a rename whose old entry was already
	// gone has nothing to do about it.
	index.Remove("never/there")
	if len(index.Packages) != 3 {
		t.Errorf("removing an absent package changed the catalogue")
	}
}
