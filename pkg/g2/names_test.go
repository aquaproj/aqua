package g2_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// The table is read out of the catalogue so that the two can't disagree about which package
// a name belongs to, or about where it is.
func TestNewNames(t *testing.T) {
	t.Parallel()
	got := g2.NewNames(&g2.Index{Packages: []*g2.IndexPackage{
		{Name: "anomalyco/opencode", ID: "1790772767", Aliases: []string{"sst/opencode"}},
		{Name: "vale-cli/vale", ID: "1790772768", Aliases: []string{"errata-ai/vale"}},
		// A package with no other name, which is nearly all of them.
		{Name: "cli/cli", ID: "1790772769"},
		// Neither of these is a name to resolve: one is empty and one is the package
		// itself. Both are reported where the catalogue is written.
		{Name: "a/b", ID: "1790772770", Aliases: []string{"", "a/b"}},
		{Name: "", Aliases: []string{"nameless"}},
		nil,
	}})
	wantIDs := map[string]string{
		"anomalyco/opencode": "1790772767",
		"vale-cli/vale":      "1790772768",
		"cli/cli":            "1790772769",
		"a/b":                "1790772770",
	}
	if diff := cmp.Diff(wantIDs, got.IDs); diff != "" {
		t.Errorf("the ids are wrong (-want +got):\n%s", diff)
	}
	wantAliases := map[string]string{
		"sst/opencode":   "anomalyco/opencode",
		"errata-ai/vale": "vale-cli/vale",
	}
	if diff := cmp.Diff(wantAliases, got.Aliases); diff != "" {
		t.Errorf("the other names are wrong (-want +got):\n%s", diff)
	}
}

// A name naming two packages is a mistake in the registry, reported where the catalogue is
// written. A table built from a catalogue that has it is still a table.
func TestNewNames_claimedTwice(t *testing.T) {
	t.Parallel()
	got := g2.NewNames(&g2.Index{Packages: []*g2.IndexPackage{
		{Name: "c/d", ID: "1", Aliases: []string{"a/b"}},
		{Name: "e/f", ID: "2", Aliases: []string{"a/b"}},
	}})
	if got.Aliases["a/b"] != "c/d" {
		t.Errorf("a/b belongs to %q, want the first package to claim it", got.Aliases["a/b"])
	}
}

// The id is what a name is resolved to, whichever of the package's names it is.
func TestNames_ID(t *testing.T) {
	t.Parallel()
	names := &g2.Names{
		IDs:     map[string]string{"anomalyco/opencode": "1790772767"},
		Aliases: map[string]string{"sst/opencode": "anomalyco/opencode"},
	}
	for _, d := range []struct {
		name string
		want string
		ok   bool
	}{
		{name: "anomalyco/opencode", want: "1790772767", ok: true},
		// The name the repository used to have is the same package.
		{name: "sst/opencode", want: "1790772767", ok: true},
		{name: "cli/cli"},
	} {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			got, ok := names.ID(d.name)
			if ok != d.ok || got != d.want {
				t.Errorf("ID(%q) is %q (%v), want %q (%v)", d.name, got, ok, d.want, d.ok)
			}
		})
	}
	// A registry with no table holds nothing as far as this is concerned, rather than
	// failing.
	var none *g2.Names
	if _, ok := none.ID("cli/cli"); ok {
		t.Error("no table answered with an id")
	}
	if got := none.Resolve("cli/cli"); got != "cli/cli" {
		t.Errorf("Resolve on no table is %q", got)
	}
}

func TestNames_Marshal(t *testing.T) {
	t.Parallel()
	got, err := (&g2.Names{
		IDs:     map[string]string{"vale-cli/vale": "1790772768", "anomalyco/opencode": "1790772767"},
		Aliases: map[string]string{"sst/opencode": "anomalyco/opencode"},
	}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "ids": {
    "anomalyco/opencode": "1790772767",
    "vale-cli/vale": "1790772768"
  },
  "aliases": {
    "sst/opencode": "anomalyco/opencode"
  }
}
`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the file is wrong (-want +got):\n%s", diff)
	}
	// And it reads back.
	read, err := g2.ReadNames(strings.NewReader(got))
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := read.ID("sst/opencode"); !ok || id != "1790772767" {
		t.Errorf("the table didn't read back: %q (%v)", id, ok)
	}
}

// refDownloader answers per branch and path, the way the repository does: a package's file is
// on that package's branch and nowhere else.
type refDownloader struct {
	files map[string]string
	calls map[string]int
}

func (d *refDownloader) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if d.calls == nil {
		d.calls = map[string]int{}
	}
	key := d.key(param.Ref, param.Path)
	d.calls[key]++
	content, ok := d.files[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return &domain.GitHubContentFile{ReadCloser: io.NopCloser(strings.NewReader(content))}, nil
}

func (d *refDownloader) key(ref, path string) string { return ref + ":" + path }

const opencodeID = "1790772767"

// opencode is a package whose repository was renamed, as the registry holds it.
func opencode() *refDownloader {
	return &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.NamesFileName: `{"ids":{"anomalyco/opencode":"` + opencodeID + `"},
			"aliases":{"sst/opencode":"anomalyco/opencode"}}`,
		g2.IDBranchName(opencodeID) + ":" + g2.Path("v1.0.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"anomalyco",
			 "repo_name":"opencode","asset":"opencode.tar.gz"}]}`,
	}}
}

// The branch is named after the package's id, so the table is what turns a name into
// somewhere to ask. The entry keeps the name aqua.yaml asked for, so that install and exec
// find it without resolving anything.
func TestClient_Resolve_alias(t *testing.T) {
	t.Parallel()
	dl := opencode()
	client := g2.New(dl, nil, "", "")

	pkgs, err := client.Resolve(t.Context(), discardLogger(), "sst/opencode", "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("got %d entries", len(pkgs))
	}
	if pkgs[0].Name != "sst/opencode" {
		t.Errorf("the entry is named %q, want the name aqua.yaml asked for", pkgs[0].Name)
	}

	// The table is read once however many packages are resolved.
	if _, err := client.Resolve(t.Context(), discardLogger(), "sst/opencode", "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if got := dl.calls[dl.key(g2.DefaultBranch, g2.NamesFileName)]; got != 1 {
		t.Errorf("read the table %d times, want once", got)
	}
}

// A cached table answers before the registry is asked, so the next run resolves without
// reading anything but the file it came for.
func TestClient_Resolve_cachedTable(t *testing.T) {
	t.Parallel()
	cache := g2.NewCache(&config.Param{CacheDir: t.TempDir()})
	first := opencode()
	if _, err := g2.New(first, cache, "", "").Resolve(t.Context(), discardLogger(),
		"sst/opencode", "v1.0.0"); err != nil {
		t.Fatal(err)
	}

	second := opencode()
	if _, err := g2.New(second, cache, "", "").Resolve(t.Context(), discardLogger(),
		"sst/opencode", "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if got := second.calls[second.key(g2.DefaultBranch, g2.NamesFileName)]; got != 0 {
		t.Errorf("read the table %d times, want never", got)
	}
}

// A table written before the package arrived doesn't know its name, and a miss is what says
// so: the registry's own table is read and the name asked again. Without that the cached copy
// would answer for a registry it is older than.
func TestClient_Resolve_staleTable(t *testing.T) {
	t.Parallel()
	cache := g2.NewCache(&config.Param{CacheDir: t.TempDir()})

	// A run before the package arrived caches a table that doesn't hold it.
	before := &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.NamesFileName: `{"ids":{"cli/cli":"1790772769"},"aliases":{}}`,
		g2.IDBranchName("1790772769") + ":" + g2.Path("v2.1.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"cli",
			 "repo_name":"cli","asset":"gh.tar.gz"}]}`,
	}}
	if _, err := g2.New(before, cache, "", "").Resolve(t.Context(), discardLogger(),
		"cli/cli", "v2.1.0"); err != nil {
		t.Fatal(err)
	}

	after := opencode()
	pkgs, err := g2.New(after, cache, "", "").Resolve(t.Context(), discardLogger(),
		"sst/opencode", "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || pkgs[0].Name != "sst/opencode" {
		t.Errorf("got %+v", pkgs)
	}
	if got := after.calls[after.key(g2.DefaultBranch, g2.NamesFileName)]; got != 1 {
		t.Errorf("read the registry's table %d times, want once", got)
	}
}

// A package the registry doesn't hold is told apart from a failure to read one, because a
// caller choosing a version asks upstream instead.
func TestClient_Resolve_notHeld(t *testing.T) {
	t.Parallel()
	dl := opencode()
	_, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(), "cli/cli", "v2.1.0")
	if !errors.Is(err, g2.ErrNoPackageBranch) {
		t.Fatalf("got %v, want the registry not holding the package", err)
	}
}

// A registry with no table holds nothing that can be resolved. The file arrives with the
// packages it describes, so an older registry, or a mirror of one, simply doesn't have it.
func TestClient_Resolve_noTable(t *testing.T) {
	t.Parallel()
	dl := &refDownloader{}
	if _, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(),
		"cli/cli", "v2.1.0"); err == nil {
		t.Fatal("want an error")
	}
}
