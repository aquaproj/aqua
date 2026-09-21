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
