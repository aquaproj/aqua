package g2_test

import (
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
		`{"source":"abc","versions":[{"version":"v1.0.0","published_at":"2026-01-01T00:00:00Z","digest":"sha256:0"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	want := &g2.Versions{Source: "abc", Versions: []*g2.Version{
		{Version: "v1.0.0", PublishedAt: "2026-01-01T00:00:00Z", Digest: "sha256:0"},
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
