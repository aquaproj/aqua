package fixcmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// fakeFinder is the configuration file the test wrote.
type fakeFinder struct {
	paths []string
}

func (f *fakeFinder) Finds(_, _ string) []string {
	return f.paths
}

// fakeReader stands in for reading the file with its imports resolved.
type fakeReader struct {
	cfg *aqua.Config
}

func (f *fakeReader) Read(_ *slog.Logger, _ string, cfg *aqua.Config) error {
	*cfg = *f.cfg
	return nil
}

// fakeNames is the registry's table of other names.
type fakeNames struct {
	aliases *g2.Aliases
	offline bool
	asked   bool
}

func (f *fakeNames) NameTable(_ context.Context, _ *slog.Logger, offline bool) *g2.Aliases {
	f.asked = true
	f.offline = offline
	return f.aliases
}

func table(aliases map[string]string) *g2.Aliases {
	return &g2.Aliases{Aliases: aliases}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aqua.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The file keeps everything about it that isn't the name: its comments, the order
// somebody chose, and the version, which is what the package is at rather than what it
// is called.
func TestRenameInFile(t *testing.T) {
	t.Parallel()
	path := write(t, `---
# the tools this repository needs
registries:
  - type: standard
    ref: v4.427.0 # renovate: depName=aquaproj/aqua-registry
packages:
  - name: sst/opencode@v0.14.1 # an agent
  - name: kubernetes/kubectl
    version: v1.34.0
  - name: sst/opencode
    registry: local
`)
	renamed, err := renameInFile(discardLogger(), path, map[string]string{
		"sst/opencode":       "anomalyco/opencode",
		"kubernetes/kubectl": "kubernetes/kubernetes/kubectl",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !renamed {
		t.Fatal("the file holds packages to rename")
	}
	want := `---
# the tools this repository needs
registries:
  - type: standard
    ref: v4.427.0 # renovate: depName=aquaproj/aqua-registry
packages:
  - name: anomalyco/opencode@v0.14.1 # an agent
  - name: kubernetes/kubernetes/kubectl
    version: v1.34.0
  - name: sst/opencode
    registry: local
`
	if diff := cmp.Diff(want, read(t, path)); diff != "" {
		t.Errorf("the file is wrong (-want +got):\n%s", diff)
	}
}

// A file that declares no packages is nothing to do rather than a failure: a
// configuration may hold only registries, or only imports.
func TestRenameInFile_noPackages(t *testing.T) {
	t.Parallel()
	path := write(t, "registries:\n  - type: standard\n    ref: v4.427.0\n")
	renamed, err := renameInFile(discardLogger(), path, map[string]string{"sst/opencode": "anomalyco/opencode"})
	if err != nil {
		t.Fatal(err)
	}
	if renamed {
		t.Error("nothing was there to rename")
	}
}

// What the table is asked about is the names in the configuration, and what comes back
// renames the file.
func TestController_Fix(t *testing.T) {
	t.Parallel()
	path := write(t, "packages:\n  - name: sst/opencode@v0.14.1\n")
	names := &fakeNames{aliases: table(map[string]string{"sst/opencode": "anomalyco/opencode"})}
	c := New(&fakeFinder{paths: []string{path}}, &fakeReader{cfg: &aqua.Config{
		Packages: []*aqua.Package{{Name: "sst/opencode@v0.14.1"}},
	}}, names)

	if err := c.Fix(t.Context(), discardLogger(), &config.Param{}, &Args{}); err != nil {
		t.Fatal(err)
	}
	if got, want := read(t, path), "packages:\n  - name: anomalyco/opencode@v0.14.1\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --check is how a job asks whether the configuration still says what the registry says:
// it reports what it found, writes nothing, and fails.
func TestController_Fix_check(t *testing.T) {
	t.Parallel()
	content := "packages:\n  - name: sst/opencode@v0.14.1\n"
	path := write(t, content)
	c := New(&fakeFinder{paths: []string{path}}, &fakeReader{cfg: &aqua.Config{
		Packages: []*aqua.Package{{Name: "sst/opencode@v0.14.1"}},
	}}, &fakeNames{aliases: table(map[string]string{"sst/opencode": "anomalyco/opencode"})})

	err := c.Fix(t.Context(), discardLogger(), &config.Param{}, &Args{Check: true})
	if !errors.Is(err, errOutOfDate) {
		t.Fatalf("a file that is out of date should be reported, got %v", err)
	}
	if got := read(t, path); got != content {
		t.Errorf("the file was written: %q", got)
	}
}

// A file already saying what the registry says is up to date, and --check passes.
func TestController_Fix_upToDate(t *testing.T) {
	t.Parallel()
	content := "packages:\n  - name: anomalyco/opencode@v0.14.1\n"
	path := write(t, content)
	c := New(&fakeFinder{paths: []string{path}}, &fakeReader{cfg: &aqua.Config{
		Packages: []*aqua.Package{{Name: "anomalyco/opencode@v0.14.1"}},
	}}, &fakeNames{aliases: table(map[string]string{"sst/opencode": "anomalyco/opencode"})})

	if err := c.Fix(t.Context(), discardLogger(), &config.Param{}, &Args{Check: true}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, path); got != content {
		t.Errorf("the file changed: %q", got)
	}
}

// A registry with no table renames nothing, which is what an older aqua-registry-g2 or a
// mirror of one looks like, and what offline looks like before anything is cached.
func TestController_Fix_noTable(t *testing.T) {
	t.Parallel()
	content := "packages:\n  - name: sst/opencode@v0.14.1\n"
	path := write(t, content)
	names := &fakeNames{}
	c := New(&fakeFinder{paths: []string{path}}, &fakeReader{cfg: &aqua.Config{
		Packages: []*aqua.Package{{Name: "sst/opencode@v0.14.1"}},
	}}, names)

	if err := c.Fix(t.Context(), discardLogger(), &config.Param{}, &Args{Offline: true}); err != nil {
		t.Fatal(err)
	}
	if !names.offline {
		t.Error("--offline should reach the table")
	}
	if got := read(t, path); got != content {
		t.Errorf("the file changed: %q", got)
	}
}

// Only the standard registry's packages are asked about. Another registry names its
// packages its own way, and the table describes this one.
func TestPackageNames(t *testing.T) {
	t.Parallel()
	got := packageNames(&aqua.Config{Packages: []*aqua.Package{
		{Name: "sst/opencode@v0.14.1"},
		{Name: "cli/cli", Registry: "standard"},
		{Name: "foo/bar", Registry: "local"},
		{Name: "imported/pkg", FilePath: "imported.yaml"},
		nil,
		{},
	}})
	want := map[string][]string{
		"sst/opencode": {""},
		"cli/cli":      {""},
		"imported/pkg": {"imported.yaml"},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the names are wrong (-want +got):\n%s", diff)
	}
}
