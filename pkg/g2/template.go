package g2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// TemplateFileName is the definition of a package served as a template, in the package's
// directory in place of the versions a generated package has.
//
// The name carries the schema's major version, the way FileName does: the fields a
// template may use are fixed for each major, because aqua reads it with encoding/json and
// an older aqua would drop a field it doesn't know without a word, resolving another asset
// than the template means and locking its checksum.
const TemplateFileName = "template-1.json"

// ErrNoTemplate is a package that has neither the version's registry.json nor a template:
// one the registry holds but hasn't generated the version of.
var ErrNoTemplate = errors.New("aqua-registry-g2 has neither a registry.json of the version nor a template of the package")

// NewTemplate is what TemplateFileName holds for a definition: the package as aqua's
// registry describes it, without what only generating versions reads (repo_id, the asset
// filters and rules, prerelease) and without Template, which says how it is checked rather
// than how it installs.
//
// The version_overrides are written as the definition has them, newest first. That is not
// how aqua-registry reads them, so a template is read with Versioned rather than as a
// registry's package.
func NewTemplate(cfg *Config) *registry.PackageInfo {
	if cfg == nil || cfg.PackageInfo == nil {
		return nil
	}
	return cfg.Copy()
}

// ParseTemplate reads a TemplateFileName.
//
// The JSON is read into the package's definition and nothing else, so that a key it has is
// never taken for one of the fields only a definition in the registry carries: encoding/json
// matches names regardless of case.
func ParseTemplate(b []byte) (*Config, error) {
	pkgInfo := &registry.PackageInfo{}
	if err := json.Unmarshal(b, pkgInfo); err != nil {
		return nil, fmt.Errorf("read the template as JSON: %w", err)
	}
	return &Config{PackageInfo: pkgInfo}, nil
}

// Versioned returns the definition of one version, ready for anything that resolves a
// registry's package.
//
// SetVersion applies the version_overrides the way this registry means them. What it returns
// still has them, and aqua's own resolution applies them again, the way aqua-registry means
// them: when the definition has no version_constraint at the top it ignores them, and
// otherwise it reads them oldest first and falls back to the top. Clearing both leaves
// nothing to apply twice, so every reader -- aqua locking and installing, ar2 testing a
// template -- resolves the version the same way.
func (c *Config) Versioned(logger *slog.Logger, v string) (*registry.PackageInfo, error) {
	pkgInfo, err := c.SetVersion(logger, v)
	if err != nil {
		return nil, err
	}
	pkgInfo.VersionConstraints = ""
	pkgInfo.VersionOverrides = nil
	return pkgInfo, nil
}

// TemplatePath is where a package's template is cached.
//
// Under the package's name, the way a registry.json is, so that a cached template is found
// again without resolving anything.
func (c *Cache) TemplatePath(repoOwner, repoName, pkgName string) string {
	return filepath.Join(c.dir, RegistryType, "github.com", repoOwner, repoName,
		EncodePackageName(pkgName), TemplateFileName)
}

// TemplateTTL is how long a cached template answers install and exec.
//
// A day, the same as the table of names. Unlike a registry.json, a template changes when its
// definition is fixed, so a copy can't be kept for good. A copy kept for a day costs a fix
// reaching a machine a day late at most, while exec, which runs on every command through
// aqua-proxy, asks the network once a day rather than every time.
const TemplateTTL = TableTTL

// cachedTemplate is the cached template, when it was written no longer than ttl ago. A ttl
// of zero takes it whatever its age.
func (c *Client) cachedTemplate(pkgName string, ttl time.Duration) *Config {
	if c.cache == nil {
		return nil
	}
	b := c.cache.ReadWithin(c.cache.TemplatePath(c.repoOwner, c.repoName, pkgName), ttl)
	if b == nil {
		return nil
	}
	tmpl, err := ParseTemplate(b)
	if err != nil {
		// A cached file that doesn't parse is a copy gone wrong. Fetching replaces it.
		return nil
	}
	return tmpl
}

// staleTemplate is the cached template whatever its age, for a run that couldn't ask the
// registry. Its age is reset, so that the next run waits a day before asking again: a
// machine that is offline would otherwise ask on every command, and each attempt is retried
// for tens of seconds before it fails.
func (c *Client) staleTemplate(logger *slog.Logger, pkgName string) *Config {
	tmpl := c.cachedTemplate(pkgName, 0)
	if tmpl == nil {
		return nil
	}
	if err := c.cache.Touch(c.cache.TemplatePath(c.repoOwner, c.repoName, pkgName)); err != nil {
		slogerr.WithError(logger, err).Debug("reset the age of the cached template")
	}
	logger.Debug("the registry couldn't be read; using the cached template")
	return tmpl
}

// fetchTemplate reads the package's template from the registry and caches it. A package
// that has none gets an error wrapping domain.ErrGitHubContentNotFound.
func (c *Client) fetchTemplate(ctx context.Context, logger *slog.Logger, pkgName string) (*Config, error) {
	dir, err := c.DirOf(ctx, logger, pkgName)
	if err != nil {
		return nil, err
	}
	file, err := c.dl.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: c.repoOwner,
		RepoName:  c.repoName,
		Ref:       DefaultBranch,
		Path:      dir + "/" + TemplateFileName,
		Public:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("download the template: %w", err)
	}
	defer file.Close()
	b, err := io.ReadAll(file.Reader())
	if err != nil {
		return nil, fmt.Errorf("read the template: %w", err)
	}
	tmpl, err := ParseTemplate(b)
	if err != nil {
		return nil, err
	}
	if c.cache != nil {
		if err := c.cache.Write(c.cache.TemplatePath(c.repoOwner, c.repoName, pkgName), b); err != nil {
			// The fetch succeeded, so failing here would throw away a good answer over
			// a copy of it.
			slogerr.WithError(logger, err).Warn("cache the template")
		}
	}
	return tmpl, nil
}
