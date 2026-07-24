//nolint:funlen
package registry_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

// newRootDir creates a root directory with the cache directory in it, and
// returns the root directory and the cache directory.
func newRootDir(t *testing.T) (string, string) {
	t.Helper()
	rootDir := t.TempDir()
	cacheDir := filepath.Join(rootDir, "registry-cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return rootDir, cacheDir
}

func TestNewCache_WithExistingFile(t *testing.T) {
	t.Parallel()
	rootDir, cacheDir := newRootDir(t)

	// Create sample cache data
	data := map[string]map[string]*registry.PackageInfo{
		"registry1": {
			"package1": {
				Name: "package1",
				Type: "github_release",
			},
		},
	}

	cacheFile := filepath.Join(cacheDir, "L3BhdGgvdG8vY29uZmlnLnlhbWw=.json")
	file, err := os.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		file.Close()
		t.Fatalf("failed to encode cache data: %v", err)
	}
	file.Close()

	// Test NewCache with existing file
	cache, err := registry.NewCache(rootDir, "/path/to/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cache == nil {
		t.Error("cache should not be nil")
	}

	// Verify data was loaded
	result := cache.Get("registry1", "package1")
	if result == nil {
		t.Error("package1 should be found")
	} else if result.GetName() != "package1" {
		t.Errorf("expected name 'package1', got %q", result.GetName())
	}
}

func TestNewCache_WithInvalidJSON(t *testing.T) {
	t.Parallel()
	rootDir, cacheDir := newRootDir(t)

	cacheFile := filepath.Join(cacheDir, "L3BhdGgvdG8vY29uZmlnLnlhbWw=.json")
	file, err := os.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	if _, err := file.WriteString("invalid json"); err != nil {
		file.Close()
		t.Fatalf("failed to write invalid json: %v", err)
	}
	file.Close()

	// Test NewCache with invalid JSON
	_, err = registry.NewCache(rootDir, "/path/to/config.yaml")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// Test cache operations without using NewCache
func TestCache_Operations(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// Create cache with valid data using direct setup
	rootDir, cacheDir := newRootDir(t)

	// Create initial cache file
	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := os.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	initialData := map[string]map[string]*registry.PackageInfo{}
	if err := json.NewEncoder(file).Encode(initialData); err != nil {
		file.Close()
		t.Fatalf("failed to encode initial data: %v", err)
	}
	file.Close()

	cache, err := registry.NewCache(rootDir, "/config.yaml")
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Test Add operation
	pkg1 := &registry.PackageInfo{Name: "package1", Type: "github_release"}
	cache.Add("registry1", pkg1)

	// Test Get operation
	result := cache.Get("registry1", "package1")
	if result == nil {
		t.Error("package1 should be found")
	} else if result.GetName() != "package1" {
		t.Errorf("expected name 'package1', got %q", result.GetName())
	}

	// Test Get non-existent package
	result = cache.Get("registry1", "nonexistent")
	if result != nil {
		t.Error("should return nil for non-existent package")
	}

	// Test Get from non-existent registry
	result = cache.Get("nonexistent", "package1")
	if result != nil {
		t.Error("should return nil for non-existent registry")
	}

	// Test Clean operation
	cache.Add("registry2", &registry.PackageInfo{Name: "package2"})
	cache.Add("registry3", &registry.PackageInfo{Name: "package3"})

	keys := map[string]map[string]struct{}{
		"registry1": {"package1": {}},
		// registry2 and registry3 not in keys, so they should be removed
	}

	cache.Clean(keys)

	// Verify cleaning results
	if result := cache.Get("registry1", "package1"); result == nil {
		t.Error("package1 should still exist in registry1")
	}
	if result := cache.Get("registry2", "package2"); result != nil {
		t.Error("registry2 should be removed")
	}
	if result := cache.Get("registry3", "package3"); result != nil {
		t.Error("registry3 should be removed")
	}

	// Test Write operation
	if err := cache.Write(); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(cacheFile); err != nil {
		t.Error("cache file should exist after write")
	}
}

func TestCache_WriteReadRoundTrip(t *testing.T) {
	t.Parallel()
	rootDir, cacheDir := newRootDir(t)

	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := os.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	initialData := map[string]map[string]*registry.PackageInfo{}
	if err := json.NewEncoder(file).Encode(initialData); err != nil {
		file.Close()
		t.Fatalf("failed to encode initial data: %v", err)
	}
	file.Close()

	// Create first cache and add data
	cache1, err := registry.NewCache(rootDir, "/config.yaml")
	if err != nil {
		t.Fatalf("failed to create cache1: %v", err)
	}

	pkg1 := &registry.PackageInfo{
		Name:      "testpkg",
		Type:      "github_release",
		RepoOwner: "owner",
		RepoName:  "repo",
	}
	cache1.Add("registry1", pkg1)

	// Write cache
	if err := cache1.Write(); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Create second cache (should read the written data)
	cache2, err := registry.NewCache(rootDir, "/config.yaml")
	if err != nil {
		t.Fatalf("failed to create cache2: %v", err)
	}

	// Verify data was loaded correctly
	result := cache2.Get("registry1", "testpkg")
	if result == nil {
		t.Error("testpkg should be found in registry1")
	} else {
		if result.GetName() != "testpkg" {
			t.Errorf("expected name 'testpkg', got %q", result.GetName())
		}
		if result.Type != "github_release" {
			t.Errorf("expected type 'github_release', got %q", result.Type)
		}
		if result.RepoOwner != "owner" {
			t.Errorf("expected repo owner 'owner', got %q", result.RepoOwner)
		}
	}
}

func TestCache_WriteError(t *testing.T) {
	t.Parallel()

	rootDir, cacheDir := newRootDir(t)

	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := os.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	initialData := map[string]map[string]*registry.PackageInfo{}
	if err := json.NewEncoder(file).Encode(initialData); err != nil {
		file.Close()
		t.Fatalf("failed to encode initial data: %v", err)
	}
	file.Close()

	cache, err := registry.NewCache(rootDir, "/config.yaml")
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Add data so that Write has something to write.
	cache.Add("registry1", &registry.PackageInfo{Name: "pkg1"})

	// Replace the cache file with a directory. Creating the file then fails,
	// which is what this test is about: the error must be reported rather than
	// swallowed. A memory-mapped filesystem made this awkward to arrange.
	if err := os.Remove(cacheFile); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(cacheFile, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := cache.Write(); err == nil {
		t.Fatal("an error must be returned when the cache file can't be created")
	}
}
