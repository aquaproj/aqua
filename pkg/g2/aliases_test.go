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

// refDownloader answers per branch and path, the way the repository does: a package's
// file is on that package's branch and nowhere else.
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

// opencode is a package whose repository was renamed, as the registry holds it.
func opencode() *refDownloader {
	return &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.AliasesFileName: `{"aliases":{"sst/opencode":"anomalyco/opencode"}}`,
		g2.BranchName("anomalyco/opencode") + ":" + g2.Path("v1.0.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"anomalyco",
			 "repo_name":"opencode","asset":"opencode.tar.gz"}]}`,
	}}
}

// The old name isn't where the file is, so the table is read and the new name tried.
// The entry keeps the name aqua.yaml asked for, so that install and exec find it
// without resolving anything.
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
	if dl.calls[dl.key(g2.DefaultBranch, g2.AliasesFileName)] != 1 {
		t.Error("the table wasn't read")
	}

	// The table is read once however many packages are resolved.
	if _, err := client.Resolve(t.Context(), discardLogger(), "sst/opencode", "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if got := dl.calls[dl.key(g2.DefaultBranch, g2.AliasesFileName)]; got != 1 {
		t.Errorf("read the table %d times, want once", got)
	}
}

// A name that answers is never checked against the table, which is nearly every name
// and the reason the table isn't fetched up front. With no copy on disk -- CI -- that
// download would be paid for each of them.
func TestClient_Resolve_noAlias(t *testing.T) {
	t.Parallel()
	dl := &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.AliasesFileName: `{"aliases":{}}`,
		g2.BranchName("cli/cli") + ":" + g2.Path("v2.1.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"cli",
			 "repo_name":"cli","asset":"gh.tar.gz"}]}`,
	}}
	if _, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(),
		"cli/cli", "v2.1.0"); err != nil {
		t.Fatal(err)
	}
	if got := dl.calls[dl.key(g2.DefaultBranch, g2.AliasesFileName)]; got != 0 {
		t.Errorf("read the table %d times, want never", got)
	}
}

// A table written before a rename doesn't hold it, so the registry's own is read and
// asked again. Without that the cache would never be corrected: it has no expiry.
func TestClient_Resolve_staleCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cache := g2.NewCache(&config.Param{CacheDir: dir})

	// A run before the rename caches a table that says nothing about the package.
	before := &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.AliasesFileName: `{"aliases":{}}`,
		g2.BranchName("sst/opencode") + ":" + g2.Path("v1.0.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"sst",
			 "repo_name":"opencode","asset":"opencode.tar.gz"}]}`,
	}}
	if _, err := g2.New(before, cache, "", "").Resolve(t.Context(), discardLogger(),
		"sst/opencode", "v1.0.0"); err != nil {
		t.Fatal(err)
	}

	// After the rename a version nothing has cached is asked for. The old name has
	// nothing, and the cached table still says there is no other.
	after := &refDownloader{files: map[string]string{
		g2.DefaultBranch + ":" + g2.AliasesFileName: `{"aliases":{"sst/opencode":"anomalyco/opencode"}}`,
		g2.BranchName("anomalyco/opencode") + ":" + g2.Path("v1.1.0"): `{"assets":[
			{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"anomalyco",
			 "repo_name":"opencode","asset":"opencode.tar.gz"}]}`,
	}}
	pkgs, err := g2.New(after, cache, "", "").Resolve(t.Context(), discardLogger(),
		"sst/opencode", "v1.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 || pkgs[0].Name != "sst/opencode" {
		t.Errorf("got %+v", pkgs)
	}
	if got := after.calls[after.key(g2.DefaultBranch, g2.AliasesFileName)]; got != 1 {
		t.Errorf("read the registry's table %d times, want once", got)
	}
}

// A cached table answers before the registry is asked, so the second run of a renamed
// package reads neither the old name nor the table.
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
	if got := second.calls[second.key(g2.DefaultBranch, g2.AliasesFileName)]; got != 0 {
		t.Errorf("read the table %d times, want never", got)
	}
	if got := second.calls[second.key(g2.BranchName("sst/opencode"), g2.Path("v1.0.0"))]; got != 0 {
		t.Errorf("asked the old name %d times, want never", got)
	}
}

// A registry with no table fails with what the name itself said, not with something
// about the table. The file arrives with the renames it describes, so an older registry
// or a mirror of one simply doesn't have it.
func TestClient_Resolve_noTable(t *testing.T) {
	t.Parallel()
	dl := &refDownloader{}
	if _, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(),
		"cli/cli", "v2.1.0"); err == nil {
		t.Fatal("want an error")
	}
}
