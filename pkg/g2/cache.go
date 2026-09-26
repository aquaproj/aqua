package g2

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

// CacheDirName is where cached registry.json files live under the cache directory.
const CacheDirName = "registries-g2"

// Cache holds files that have already been fetched.
//
// A registry.json never expires. It describes one version of one package and ar2 writes
// it once, so a cached copy can't go stale the way a registry covering every package
// could. --no-cache is therefore for the case where a file was written wrong, not for
// the case where it is out of date.
//
// The table of other names can go stale, since a package is renamed after it is
// written. It is safe to keep anyway because it is only ever read to find a package
// that wasn't where its name said: a stale copy costs the fetch that replaces it, and
// never answers with a name it doesn't hold.
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

// AliasesPath is where the table of other names is cached.
//
// Beside the branches rather than inside one: it is the registry as a whole, the same
// way it sits at the root of the default branch rather than on a package's branch.
func (c *Cache) AliasesPath(repoOwner, repoName string) string {
	return filepath.Join(c.dir, RegistryType, "github.com", repoOwner, repoName, AliasesFileName)
}

// Read returns the cached file, or nil when there is none.
//
// A cache that can't be read is not an error. The file is a copy of something that
// can be fetched again, so the caller carries on and fetches it.
func (c *Cache) Read(path string) []byte {
	return c.ReadWithin(path, 0)
}

// ReadWithin returns the cached file when it was written no longer than ttl ago, and
// nil otherwise. A ttl of zero or less accepts it whatever its age.
//
// An age is asked for by whoever is asking the file a question a missing answer looks
// like an answer to. Reading it to find a package that wasn't where its name said can
// take any copy, because a copy that doesn't hold the rename simply doesn't answer;
// asking whether a name has been renamed can't tell a stale copy from a name that
// never was.
func (c *Cache) ReadWithin(path string, ttl time.Duration) []byte {
	if c.skipRead {
		return nil
	}
	if ttl > 0 {
		stat, err := os.Stat(path)
		if err != nil || time.Since(stat.ModTime()) > ttl {
			return nil
		}
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
