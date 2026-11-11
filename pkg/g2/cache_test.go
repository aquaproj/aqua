package g2_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// A cached file sits under the name it was asked for rather than under the branch it came
// from, which is named after the package's id: a file that has been read once is found again
// without resolving anything.
func TestCache_Path(t *testing.T) {
	t.Parallel()
	cache := g2.NewCache(&config.Param{CacheDir: "/cache"})
	want := filepath.Join("/cache", "registries-g2", "github_content", "github.com",
		"aquaproj", "aqua-registry-g2", "cli_2fcli", "versions", "v2.1.0", "registry-1.json")
	if diff := cmp.Diff(want, cache.Path("aquaproj", "aqua-registry-g2", "cli/cli", "v2.1.0")); diff != "" {
		t.Errorf("the path is wrong (-want +got):\n%s", diff)
	}
}

func TestCache_WriteAndRead(t *testing.T) {
	t.Parallel()
	cache := g2.NewCache(&config.Param{CacheDir: t.TempDir()})
	path := cache.Path("aquaproj", "aqua-registry-g2", "cli/cli", "v2.1.0")

	// Nothing is cached yet, and that isn't an error: the caller fetches instead.
	if b := cache.Read(path); b != nil {
		t.Fatalf("got %q, want nothing", b)
	}
	if err := cache.Write(path, []byte(`{"assets":[]}`)); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"assets":[]}`, string(cache.Read(path))); diff != "" {
		t.Errorf("the cached file is wrong (-want +got):\n%s", diff)
	}
}

// --no-cache has to miss rather than skip the cache entirely, so that the refetched
// file replaces the one that was there.
func TestCache_noCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	written := g2.NewCache(&config.Param{CacheDir: dir})
	path := written.Path("aquaproj", "aqua-registry-g2", "cli/cli", "v2.1.0")
	if err := written.Write(path, []byte("cached")); err != nil {
		t.Fatal(err)
	}

	cache := g2.NewCache(&config.Param{CacheDir: dir, NoCache: true})
	if b := cache.Read(path); b != nil {
		t.Errorf("got %q, want nothing", b)
	}
	if err := cache.Write(path, []byte("fetched")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("fetched", string(b)); diff != "" {
		t.Errorf("the cached file is wrong (-want +got):\n%s", diff)
	}
}

// Without somewhere to put it there is no cache, and reading still works.
func TestNewCache_noDir(t *testing.T) {
	t.Parallel()
	if cache := g2.NewCache(&config.Param{}); cache != nil {
		t.Errorf("got %+v, want nothing", cache)
	}
}
