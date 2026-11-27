package g2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/resolve"
)

// RegistryType is the registry type recorded in a lock entry resolved from g2.
//
// g2 is read as files in a GitHub repository, the same way a github_content registry
// is, so an entry from it names that type rather than inventing one. What makes it g2
// is the repository, which the entry also records.
const RegistryType = "github_content"

// LockPackages converts a package version's registry.json into lock file entries,
// one per environment.
//
// The conversion is a copy rather than a transformation: registry.json is already
// resolved for each environment, and a lock entry is the same thing plus the package
// name, the version and where it came from. Anything computed here would be a second
// place for the two formats to disagree.
func (c *Client) LockPackages(reg *Registry, pkgName, version string) []*lockfile.Package {
	pkgs := make([]*lockfile.Package, 0, len(reg.Assets))
	for _, asset := range reg.Assets {
		pkgs = append(pkgs, &lockfile.Package{
			Name:    pkgName,
			Version: version,
			Registry: &lockfile.Registry{
				Type:      RegistryType,
				RepoOwner: c.repoOwner,
				RepoName:  c.repoName,
			},
			OS:                         asset.OS,
			Arch:                       asset.Arch,
			Variants:                   asset.Variants,
			LinkedLibc:                 asset.LinkedLibc,
			Rosetta2:                   asset.Rosetta2,
			WindowsARMEmulation:        asset.WindowsARMEmulation,
			Checksum:                   asset.Checksum,
			ChecksumAlgorithm:          asset.ChecksumAlgorithm,
			Type:                       asset.Type,
			RepoOwner:                  asset.RepoOwner,
			RepoName:                   asset.RepoName,
			RepoID:                     asset.RepoID,
			Host:                       asset.Host,
			Asset:                      asset.Asset,
			URL:                        asset.URL,
			Format:                     asset.Format,
			Path:                       asset.Path,
			Crate:                      asset.Crate,
			Cargo:                      asset.Cargo,
			Files:                      lockFiles(asset.Files),
			Private:                    asset.Private,
			Cosign:                     asset.Cosign,
			GitHubArtifactAttestations: asset.GitHubArtifactAttestations,
			Minisign:                   asset.Minisign,
			Packslip:                   asset.Packslip,
			SLSAProvenance:             asset.SLSAProvenance,
		})
	}
	return pkgs
}

func lockFiles(files []*File) []*lockfile.File {
	if len(files) == 0 {
		return nil
	}
	out := make([]*lockfile.File, len(files))
	for i, f := range files {
		out[i] = &lockfile.File{
			Name: f.Name,
			Src:  f.Src,
		}
	}
	return out
}

// Resolved is a package version resolved for every environment it supports.
type Resolved struct {
	Packages []*lockfile.Package
	// PkgInfo is the definition of the version the entries were expanded from, for a
	// package served as a template. Its entries carry no checksum, since nothing verified
	// them ahead of time, and fetching the checksums needs it. It is nil for a generated
	// registry.json, whose entries carry their checksums already.
	PkgInfo *registry.PackageInfo
}

// Resolve returns the lock file entries for one package version.
//
// It is the g2 arm of the lock update command's resolution order. The version's
// registry.json answers when the registry generated one; otherwise the package's template
// does, expanded for supportedEnvs. Only a package served as a template has one, so a
// version of any other package never falls through to it.
//
// The template is read from the registry every time rather than from the cache: its
// definition can be fixed after it is written, and what is locked is what it says now.
func (c *Client) Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*Resolved, error) {
	// Both names reach the package: the table maps a name a repository used to have to
	// the one it has now, and both of them to the same id. So resolving is the
	// fetch's business, and what is left here is saying that the configuration asks for
	// a name the registry no longer calls it.
	resolved, err := c.definitions(ctx, logger, pkgName, version, supportedEnvs, true)
	if err != nil {
		return nil, err
	}
	c.sayRenamed(logger, pkgName, c.cachedNames().Resolve(pkgName))
	return resolved, nil
}

// Definitions returns the definitions of one package version, the way Resolve does, for
// installing a locked package.
//
// It says nothing about a rename. Install and exec run it, and exec runs on every command
// through aqua-proxy: a warning there would be printed beside the output of the tool
// itself, every time. "aqua lock update" is where the rename is said.
//
// A template cached within TemplateTTL answers without the network, and so does one of any
// age when the registry can't be read: a package installed already keeps running offline,
// as a generated one does.
func (c *Client) Definitions(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string) (*Resolved, error) {
	return c.definitions(ctx, logger, pkgName, version, supportedEnvs, false)
}

func (c *Client) definitions(ctx context.Context, logger *slog.Logger, pkgName, version string, supportedEnvs []string, fresh bool) (*Resolved, error) {
	if !fresh {
		if reg := c.cachedRegistry(logger, pkgName, version); reg != nil {
			return &Resolved{Packages: c.LockPackages(reg, pkgName, version)}, nil
		}
		if tmpl := c.cachedTemplate(pkgName, TemplateTTL); tmpl != nil {
			return c.expand(logger, tmpl, pkgName, version, supportedEnvs)
		}
	}
	reg, err := c.Get(ctx, logger, pkgName, version)
	if err == nil {
		// The entry is named the way aqua.yaml asked for it, not the way the registry
		// holds it. That name is what ties the entry to the configuration, so install
		// and exec find it without resolving anything -- which is what keeps the table
		// off the path that runs on every command.
		return &Resolved{Packages: c.LockPackages(reg, pkgName, version)}, nil
	}
	if !errors.Is(err, domain.ErrGitHubContentNotFound) {
		return c.fallBack(logger, pkgName, version, supportedEnvs, fresh, err)
	}
	tmpl, tmplErr := c.fetchTemplate(ctx, logger, pkgName)
	if tmplErr == nil {
		return c.expand(logger, tmpl, pkgName, version, supportedEnvs)
	}
	if errors.Is(tmplErr, domain.ErrGitHubContentNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrNoTemplate, err)
	}
	return c.fallBack(logger, pkgName, version, supportedEnvs, fresh, tmplErr)
}

// fallBack answers from a cached template of any age when the registry couldn't be read,
// for install and exec. The lock update command fails instead: what it records has to be
// what the registry says now.
func (c *Client) fallBack(logger *slog.Logger, pkgName, version string, supportedEnvs []string, fresh bool, err error) (*Resolved, error) {
	if !fresh {
		if tmpl := c.staleTemplate(logger, pkgName); tmpl != nil {
			return c.expand(logger, tmpl, pkgName, version, supportedEnvs)
		}
	}
	return nil, err
}

// cachedRegistry is the version's cached registry.json, or nil.
func (c *Client) cachedRegistry(logger *slog.Logger, pkgName, version string) *Registry {
	if c.cache == nil {
		return nil
	}
	cachePath := c.cache.Path(c.repoOwner, c.repoName, pkgName, version)
	b := c.cache.Read(cachePath)
	if b == nil {
		return nil
	}
	reg, err := parseRegistry(b)
	if err != nil {
		logger.Debug("the cached registry.json is broken", "cache_path", cachePath)
		return nil
	}
	return reg
}

// expand resolves a template for one version, for every environment in supportedEnvs (every
// environment the definition supports when it is empty).
func (c *Client) expand(logger *slog.Logger, tmpl *Config, pkgName, version string, supportedEnvs []string) (*Resolved, error) {
	versioned, err := tmpl.Versioned(logger, version)
	if err != nil {
		return nil, fmt.Errorf("apply version_overrides to the template: %w", err)
	}
	pkgs, err := resolve.Resolve(logger, &resolve.Param{
		PkgName:       pkgName,
		Version:       version,
		PkgInfo:       versioned,
		SupportedEnvs: supportedEnvs,
	})
	if err != nil {
		return nil, fmt.Errorf("expand the template: %w", err)
	}
	for _, pkg := range pkgs {
		pkg.Registry = &lockfile.Registry{
			Type:      RegistryType,
			RepoOwner: c.repoOwner,
			RepoName:  c.repoName,
		}
	}
	return &Resolved{Packages: pkgs, PkgInfo: versioned}, nil
}

// sayRenamed tells whoever is watching that the package has another name now.
//
// Said on every run rather than once, because nothing here changes aqua.yaml: the file
// keeps asking for the old name until a person decides otherwise, and a warning that
// stopped would leave that decision unmade and unmentioned.
func (c *Client) sayRenamed(logger *slog.Logger, asked, found string) {
	if asked == found {
		return
	}
	logger.Warn("the registry knows this package by another name now",
		"package", asked, "renamed_to", found)
}

// NewDefault creates a Client reading aqua-registry-g2.
//
// The dependency injection graph has no string to pass, so the repository is fixed
// here rather than plumbed through as configuration nobody sets.
func NewDefault(dl Downloader, cache *Cache) *Client {
	return New(dl, cache, "", "")
}

// NewRegistry builds a package version's registry.json from resolved entries.
//
// It is the reverse of LockPackages, and exists so that what generates aqua-registry-g2
// and what reads it agree on the format by construction rather than by two
// descriptions of it being kept in step. The package name, the version and the
// registry are dropped: the path already says which package and which
// version this is, and the registry is the repository the file sits in.
func NewRegistry(pkgs []*lockfile.Package) *Registry {
	reg := &Registry{Assets: make([]*Asset, 0, len(pkgs))}
	for _, pkg := range pkgs {
		reg.Assets = append(reg.Assets, &Asset{
			OS:                         pkg.OS,
			Arch:                       pkg.Arch,
			Variants:                   pkg.Variants,
			LinkedLibc:                 pkg.LinkedLibc,
			Rosetta2:                   pkg.Rosetta2,
			WindowsARMEmulation:        pkg.WindowsARMEmulation,
			Type:                       pkg.Type,
			RepoOwner:                  pkg.RepoOwner,
			RepoName:                   pkg.RepoName,
			RepoID:                     pkg.RepoID,
			Host:                       pkg.Host,
			Asset:                      pkg.Asset,
			URL:                        pkg.URL,
			Format:                     pkg.Format,
			Path:                       pkg.Path,
			Crate:                      pkg.Crate,
			Cargo:                      pkg.Cargo,
			Checksum:                   pkg.Checksum,
			ChecksumAlgorithm:          pkg.ChecksumAlgorithm,
			Files:                      registryFiles(pkg.Files),
			Private:                    pkg.Private,
			Cosign:                     pkg.Cosign,
			GitHubArtifactAttestations: pkg.GitHubArtifactAttestations,
			Minisign:                   pkg.Minisign,
			Packslip:                   pkg.Packslip,
			SLSAProvenance:             pkg.SLSAProvenance,
		})
	}
	return reg
}

func registryFiles(files []*lockfile.File) []*File {
	if len(files) == 0 {
		return nil
	}
	out := make([]*File, len(files))
	for i, f := range files {
		out[i] = &File{
			Name: f.Name,
			Src:  f.Src,
		}
	}
	return out
}
