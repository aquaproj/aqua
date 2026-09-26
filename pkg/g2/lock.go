package g2

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/lockfile"
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
			Checksum:                   asset.Checksum,
			ChecksumAlgorithm:          asset.ChecksumAlgorithm,
			Type:                       asset.Type,
			RepoOwner:                  asset.RepoOwner,
			RepoName:                   asset.RepoName,
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

// Resolve returns the lock file entries for one package version.
//
// It is the g2 arm of the lock update command's resolution order. The registry cache
// and aqua-registry arms answer the same question, so they get the same shape: the
// caller tries each in turn and keeps the first that answers.
func (c *Client) Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error) {
	// The name is part of the path, so a package whose repository was renamed isn't
	// where its old name says it is. aqua-registry keeps the old name as an alias, and
	// somebody whose aqua.yaml still says it has no other way in.
	//
	// The cached table answers first, which costs nothing. Without one the name is
	// used as it is: a table downloaded to say a name was never renamed would be
	// downloaded for nearly every name.
	name := c.cachedAliases().Resolve(pkgName)
	c.sayRenamed(logger, pkgName, name)
	reg, err := c.Get(ctx, logger, name, version)
	if err == nil {
		// The entry is named the way aqua.yaml asked for it, not the way the registry
		// holds it. That name is what ties the entry to the configuration, so install
		// and exec find it without resolving anything -- which is what keeps the table
		// off the path that runs on every command.
		return c.LockPackages(reg, pkgName, version), nil
	}

	// It wasn't there. Either the package was renamed and nothing here knew -- no
	// cached table, or one written before the rename -- or the name is wrong and the
	// version isn't published. The registry's own table tells those apart, and is
	// cached so that the next run knows.
	fresh := c.fetchAliases(ctx, logger).Resolve(pkgName)
	if fresh == name {
		return nil, err
	}
	c.sayRenamed(logger, pkgName, fresh)
	reg, aliasErr := c.Get(ctx, logger, fresh, version)
	if aliasErr != nil {
		// What the caller asked for is the name in its configuration, so that is the
		// failure to report. The other is about a name it never mentioned.
		return nil, err
	}
	return c.LockPackages(reg, pkgName, version), nil
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
// registry are dropped: the branch and the path already say which package and which
// version this is, and the registry is the repository the file sits in.
func NewRegistry(pkgs []*lockfile.Package) *Registry {
	reg := &Registry{Assets: make([]*Asset, 0, len(pkgs))}
	for _, pkg := range pkgs {
		reg.Assets = append(reg.Assets, &Asset{
			OS:                         pkg.OS,
			Arch:                       pkg.Arch,
			Variants:                   pkg.Variants,
			LinkedLibc:                 pkg.LinkedLibc,
			Type:                       pkg.Type,
			RepoOwner:                  pkg.RepoOwner,
			RepoName:                   pkg.RepoName,
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
