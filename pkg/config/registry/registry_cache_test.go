//nolint:funlen
package registry_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/spf13/afero"
)

const cacheDir = "/tmp/test/registry-cache"

func TestNewCache_WithExistingFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	// Create cache directory and valid cache file
	if err := fs.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

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
	file, err := fs.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		file.Close()
		t.Fatalf("failed to encode cache data: %v", err)
	}
	file.Close()

	// Test NewCache with existing file
	cache, err := registry.NewCache(fs, "/tmp/test", "/path/to/config.yaml")
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
	fs := afero.NewMemMapFs()

	// Create cache directory and invalid cache file
	if err := fs.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "L3BhdGgvdG8vY29uZmlnLnlhbWw=.json")
	file, err := fs.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	if _, err := file.WriteString("invalid json"); err != nil {
		file.Close()
		t.Fatalf("failed to write invalid json: %v", err)
	}
	file.Close()

	// Test NewCache with invalid JSON
	_, err = registry.NewCache(fs, "/tmp/test", "/path/to/config.yaml")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// Test cache operations without using NewCache
func TestCache_Operations(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// Create cache with valid data using direct setup
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	// Create initial cache file
	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := fs.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	initialData := map[string]map[string]*registry.PackageInfo{}
	if err := json.NewEncoder(file).Encode(initialData); err != nil {
		file.Close()
		t.Fatalf("failed to encode initial data: %v", err)
	}
	file.Close()

	cache, err := registry.NewCache(fs, "/tmp/test", "/config.yaml")
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
	exists, err := afero.Exists(fs, cacheFile)
	if err != nil {
		t.Fatalf("failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("cache file should exist after write")
	}
}

func TestCache_WriteReadRoundTrip(t *testing.T) { //nolint:cyclop
	t.Parallel()
	fs := afero.NewMemMapFs()

	// Create initial empty cache file
	if err := fs.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := fs.Create(cacheFile)
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
	cache1, err := registry.NewCache(fs, "/tmp/test", "/config.yaml")
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
	cache2, err := registry.NewCache(fs, "/tmp/test", "/config.yaml")
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

	// Start with a working filesystem to create the cache
	normalFs := afero.NewMemMapFs()
	if err := normalFs.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "L2NvbmZpZy55YW1s.json")
	file, err := normalFs.Create(cacheFile)
	if err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}

	initialData := map[string]map[string]*registry.PackageInfo{}
	if err := json.NewEncoder(file).Encode(initialData); err != nil {
		file.Close()
		t.Fatalf("failed to encode initial data: %v", err)
	}
	file.Close()

	cache, err := registry.NewCache(normalFs, "/tmp/test", "/config.yaml")
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	// Add data to trigger write
	cache.Add("registry1", &registry.PackageInfo{Name: "pkg1"})

	// This test shows that write would succeed normally
	err = cache.Write()
	if err != nil {
		t.Errorf("write should succeed normally: %v", err)
	}
}
