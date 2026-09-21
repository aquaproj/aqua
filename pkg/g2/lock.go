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
			Cosign:                     asset.Cosign,
			GitHubArtifactAttestations: asset.GitHubArtifactAttestations,
			Minisign:                   asset.Minisign,
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
	reg, err := c.Get(ctx, logger, pkgName, version)
	if err != nil {
		return nil, err
	}
	return c.LockPackages(reg, pkgName, version), nil
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
			Cosign:                     pkg.Cosign,
			GitHubArtifactAttestations: pkg.GitHubArtifactAttestations,
			Minisign:                   pkg.Minisign,
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
