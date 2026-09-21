package g2

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

// CacheDirName is where cached registry.json files live under the cache directory.
const CacheDirName = "registries-g2"

// Cache holds registry.json files that have already been fetched.
//
// They never expire. A registry.json describes one version of one package and ar2
// writes it once, so a cached copy can't go stale the way a registry covering every
// package could. --no-cache is therefore for the case where a file was written wrong,
// not for the case where it is out of date.
type Cache struct {
	dir string
	// skipRead makes a lookup miss even when the file is there, so that --no-cache
	// refetches and overwrites rather than leaving the bad copy in place.
	skipRead bool
}

// NewCache creates a Cache. It returns nil when there is nowhere to put one, which
// disables caching rather than failing: fetching still works.
func NewCache(param *config.Param) *Cache {
	if param.CacheDir == "" {
		return nil
	}
	return &Cache{
		dir:      filepath.Join(param.CacheDir, CacheDirName),
		skipRead: param.NoCache,
	}
}

// Path is where one package version's registry.json is cached.
//
// It mirrors the repository's own layout, down to the branch name, so that a cached
// file can be matched to what it came from by looking at it.
func (c *Cache) Path(repoOwner, repoName, pkgName, version string) string {
	return filepath.Join(c.dir, RegistryType, "github.com", repoOwner, repoName,
		BranchName(pkgName), VersionDir, version, FileName)
}

// Read returns the cached file, or nil when there is none.
//
// A cache that can't be read is not an error. The file is a copy of something that
// can be fetched again, so the caller carries on and fetches it.
func (c *Cache) Read(path string) []byte {
	if c.skipRead {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return b
}

// Write stores the file.
func (c *Cache) Write(path string, b []byte) error {
	if err := osfile.MkdirAll(filepath.Dir(path)); err != nil {
		return fmt.Errorf("create the cache directory: %w", err)
	}
	if err := os.WriteFile(path, b, filePerm); err != nil {
		return fmt.Errorf("write the cache file: %w", err)
	}
	return nil
}

// filePerm is the mode a cached file is created with.
const filePerm = 0o644
