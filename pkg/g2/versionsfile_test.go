package g2_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// The newest release first, and a version with no release to date it after every version
// that has one.
func TestVersions_Sort(t *testing.T) {
	t.Parallel()
	versions := &g2.Versions{Versions: []*g2.Version{
		{Version: "v1.0.0", PublishedAt: "2026-01-01T00:00:00Z"},
		{Version: "nightly"},
		{Version: "v2.0.0", PublishedAt: "2026-06-01T00:00:00Z"},
		{Version: "v1.0.1", PublishedAt: "2026-01-01T00:00:00Z"},
	}}
	versions.Sort()
	if diff := cmp.Diff([]string{"v2.0.0", "v1.0.0", "v1.0.1", "nightly"}, versions.Tags()); diff != "" {
		t.Error(diff)
	}
}

// A list read back is the list that was written.
func TestReadVersions(t *testing.T) {
	t.Parallel()
	got, err := g2.ReadVersions(strings.NewReader(
		`{"source":"abc","versions":[{"version":"v1.0.0","published_at":"2026-01-01T00:00:00Z","digest":"sha256:0"},{"version":"v1.1.0-rc1","prerelease":true}]}`))
	if err != nil {
		t.Fatal(err)
	}
	want := &g2.Versions{Source: "abc", Versions: []*g2.Version{
		{Version: "v1.0.0", PublishedAt: "2026-01-01T00:00:00Z", Digest: "sha256:0"},
		{Version: "v1.1.0-rc1", Prerelease: true},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}
}

// Nothing is read from a file that isn't JSON.
func TestReadVersions_notJSON(t *testing.T) {
	t.Parallel()
	if _, err := g2.ReadVersions(strings.NewReader("<html>")); err == nil {
		t.Fatal("an error should be returned")
	}
}

// A list nobody has read yet answers for itself rather than panicking.
func TestVersions_nil(t *testing.T) {
	t.Parallel()
	var versions *g2.Versions
	versions.Sort()
	if got := versions.Tags(); got != nil {
		t.Errorf("the tags are %v", got)
	}
}

// The list is read from the package's directory on the default branch, the name resolved
// through the table.
func TestClient_ReadVersionsFile(t *testing.T) {
	t.Parallel()
	dl := opencode()
	dl.files[g2.DefaultBranch+":"+g2.PackageDir(opencodeID)+"/"+g2.VersionsFileName] = `{"versions":[
		{"version":"v1.1.0","prerelease":true},{"version":"v1.0.0"}]}`
	versions, err := g2.New(dl, nil, "", "").ReadVersionsFile(t.Context(), discardLogger(), "sst/opencode")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]*g2.Version{{Version: "v1.1.0", Prerelease: true}, {Version: "v1.0.0"}}, versions.Versions); diff != "" {
		t.Error(diff)
	}
}

// A package the registry doesn't hold is told apart from a file that couldn't be read.
func TestClient_ReadVersionsFile_notHeld(t *testing.T) {
	t.Parallel()
	c := g2.New(opencode(), nil, "", "")
	if _, err := c.ReadVersionsFile(t.Context(), discardLogger(), "cli/cli"); !errors.Is(err, g2.ErrNoPackageBranch) {
		t.Errorf("got %v, want the registry not holding the package", err)
	}
	if _, err := c.ReadVersionsFile(t.Context(), discardLogger(), "sst/opencode"); err == nil || errors.Is(err, g2.ErrNoPackageBranch) {
		t.Errorf("got %v, want a failure to read the file", err)
	}
}

// A table that can't be read says nothing about whether the registry holds the package,
// so it isn't taken for the registry not holding it.
func TestClient_DirOf_tableUnreadable(t *testing.T) {
	t.Parallel()
	_, err := g2.New(&refDownloader{}, nil, "", "").DirOf(t.Context(), discardLogger(), "cli/cli")
	if err == nil || errors.Is(err, g2.ErrNoPackageBranch) {
		t.Errorf("got %v, want a failure to read the table", err)
	}
}
