package g2_test

import (
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// The table is inverted from the catalogue so that the two can't disagree about which
// package an alias names.
func TestNewAliases(t *testing.T) {
	t.Parallel()
	got := g2.NewAliases(&g2.Index{Packages: []*g2.IndexPackage{
		{Name: "anomalyco/opencode", Aliases: []string{"sst/opencode"}},
		{Name: "vale-cli/vale", Aliases: []string{"errata-ai/vale"}},
		// A package with none, which is nearly all of them.
		{Name: "cli/cli"},
		// Neither of these is a name to resolve: one is empty and one is the package
		// itself. Both are reported where the catalogue is written.
		{Name: "a/b", Aliases: []string{"", "a/b"}},
		{Name: "", Aliases: []string{"nameless"}},
		nil,
	}})
	want := map[string]string{
		"sst/opencode":   "anomalyco/opencode",
		"errata-ai/vale": "vale-cli/vale",
	}
	if diff := cmp.Diff(want, got.Aliases); diff != "" {
		t.Errorf("the table is wrong (-want +got):\n%s", diff)
	}
}

// An alias naming two packages is a mistake in the registry, reported where the
// catalogue is written. A table built from a catalogue that has it is still a table.
func TestNewAliases_claimedTwice(t *testing.T) {
	t.Parallel()
	got := g2.NewAliases(&g2.Index{Packages: []*g2.IndexPackage{
		{Name: "c/d", Aliases: []string{"a/b"}},
		{Name: "e/f", Aliases: []string{"a/b"}},
	}})
	if got.Aliases["a/b"] != "c/d" {
		t.Errorf("a/b resolves to %q, want the first package to claim it", got.Aliases["a/b"])
	}
}

func TestAliases_Resolve(t *testing.T) {
	t.Parallel()
	aliases := &g2.Aliases{Aliases: map[string]string{"sst/opencode": "anomalyco/opencode"}}
	for _, d := range []struct{ name, want string }{
		{name: "sst/opencode", want: "anomalyco/opencode"},
		// A name that is no alias is the answer itself, so a caller resolves every
		// name it has without asking whether it needs to.
		{name: "cli/cli", want: "cli/cli"},
	} {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := aliases.Resolve(d.name); got != d.want {
				t.Errorf("Resolve(%q) is %q, want %q", d.name, got, d.want)
			}
		})
	}
	// A registry with no table resolves every name to itself rather than failing.
	var none *g2.Aliases
	if got := none.Resolve("cli/cli"); got != "cli/cli" {
		t.Errorf("Resolve on no table is %q", got)
	}
}

func TestAliases_Marshal(t *testing.T) {
	t.Parallel()
	got, err := (&g2.Aliases{Aliases: map[string]string{
		"sst/opencode":   "anomalyco/opencode",
		"errata-ai/vale": "vale-cli/vale",
	}}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "aliases": {
    "errata-ai/vale": "vale-cli/vale",
    "sst/opencode": "anomalyco/opencode"
  }
}
`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the file is wrong (-want +got):\n%s", diff)
	}
	// And it reads back.
	read, err := g2.ReadAliases(strings.NewReader(got))
	if err != nil {
		t.Fatal(err)
	}
	if read.Resolve("sst/opencode") != "anomalyco/opencode" {
		t.Error("the table didn't read back")
	}
}
