package genrgst

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// gitLabHost is the instance a package name can say a project is on.
//
// Only this one, and it is the default a gitlab_release package has rather than a second
// spelling of it: a package on another GitLab instance is a definition somebody writes,
// and its host is the one thing such a definition has to say.
const gitLabHost = registry.DefaultGitLabHost

// getGitLabPackageInfo generates the definition of a project on gitlab.com.
//
// The name says where the project is, the way crates.io/<crate> says a crate: what follows
// the host is the project's path, whose last segment is the project and whose earlier ones
// are the namespace it is in. Nothing else here is GitLab's own -- the asset naming is
// read off the release's asset list by the same inference every other type goes through.
func (c *Controller) getGitLabPackageInfo(ctx context.Context, logger *slog.Logger, pkgName, version string, limit int, cfg *Config) (*registry.PackageInfo, []string) {
	project := strings.TrimPrefix(pkgName, gitLabHost+"/")
	owner, repo, found := strings.CutLast(project, "/")
	pkgInfo := &registry.PackageInfo{
		Type:          pkgTypeGitLabRelease,
		VersionPrefix: cfg.VersionPrefix,
	}
	if cfg.VersionFilter != nil {
		pkgInfo.VersionFilter = cfg.VersionFilter.Source().String()
	}
	if !found {
		// A project needs a namespace, which is what a name of one segment is missing.
		pkgInfo.Name = pkgName
		return pkgInfo, nil
	}
	// No name and no host: the package is named <host>/<namespace>/<project> and is on
	// gitlab.com unless it says otherwise, which is what this definition means anyway.
	pkgInfo.RepoOwner = owner
	pkgInfo.RepoName = repo

	// What the project is and what it released are two answers neither of which needs the
	// other, so they are asked for at once. Somebody is waiting for the definition.
	description := make(chan string, 1)
	go func() {
		defer close(description)
		d, err := c.gitlab.GetDescription(ctx, gitLabHost, project)
		if err != nil {
			slogerr.WithError(logger, err).Warn("get the project", "project", project)
			return
		}
		description <- d
	}()

	versions := c.gitLabVersions(ctx, logger, pkgInfo, pkgName, project, version, limit, cfg)
	// Empty if the project wasn't found, which is a definition without a description
	// rather than no definition.
	pkgInfo.Description = <-description
	heldToGitLab(pkgInfo)
	return pkgInfo, versions
}

// gitLabVersions reads the releases a definition is built from, and says which they were.
//
// One release, or the one a version names, is a definition of its own; more than one is a
// definition with version_overrides, the same as for a GitHub repository.
func (c *Controller) gitLabVersions(ctx context.Context, logger *slog.Logger, pkgInfo *registry.PackageInfo, pkgName, project, version string, limit int, cfg *Config) []string {
	if limit != 1 && version == "" {
		return c.versionOverrides(logger, pkgName, pkgInfo,
			c.gitLabReleases(ctx, logger, project, limit, cfg))
	}

	release, err := c.gitlab.GetRelease(ctx, gitLabHost, project, version)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get the release", "project", project)
		return []string{version}
	}
	logger.Debug("got the release", "version", release.TagName)

	assetNames := make([]string, 0, len(release.Assets))
	for _, assetName := range release.Assets {
		if excludeAsset(logger, assetName, cfg) {
			continue
		}
		assetNames = append(assetNames, assetName)
	}
	logger.Debug("got assets", "num_of_assets", len(assetNames))

	c.patchRelease(logger, pkgInfo, pkgName, release.TagName, assetNames)
	return []string{release.TagName}
}

// gitLabReleases are the releases of a project, with what each publishes.
//
// The asset list comes with the release rather than per release, which is one request for
// a page of them: GitLab answers with a release's links inside the release.
func (c *Controller) gitLabReleases(ctx context.Context, logger *slog.Logger, project string, limit int, cfg *Config) []*Release {
	releases := []*Release{}
	for page := 1; page <= maxPages; page++ {
		listed, err := c.gitlab.ListReleases(ctx, gitLabHost, project, page, maxPerPage)
		if err != nil {
			slogerr.WithError(logger, err).Warn("list releases", "project", project)
			return releases
		}
		if len(listed) == 0 {
			return releases
		}
		for _, listedRelease := range listed {
			if listedRelease.Draft {
				// Dated in the future: its assets are not published yet, so what
				// it says they are named is not what they will be.
				continue
			}
			if excludeVersion(logger, listedRelease.TagName, cfg) {
				continue
			}
			releases = append(releases, gitLabRelease(logger, listedRelease, cfg))
			if limit > 0 && len(releases) >= limit {
				return releases
			}
		}
	}
	return releases
}

// gitLabRelease is one release as the definition is built from it.
func gitLabRelease(logger *slog.Logger, listed *forge.Release, cfg *Config) *Release {
	v, prefix, err := versiongetter.GetVersionAndPrefix(listed.TagName)
	if err != nil {
		slogerr.WithError(logger, err).Warn("parse a tag as semver", "tag_name", listed.TagName)
	}
	release := &Release{
		Tag:           listed.TagName,
		Version:       v,
		VersionPrefix: prefix,
	}
	for _, assetName := range listed.Assets {
		if excludeAsset(logger, assetName, cfg) {
			continue
		}
		release.assets = append(release.assets, assetName)
	}
	return release
}

// How far the releases of a project are read: a page at a time, and no further than an
// answer that is empty. The page is GitLab's largest, and 50 pages of releases is already
// far more than a definition is ever built from.
const (
	maxPerPage = 100
	maxPages   = 50
)

// heldToGitLab drops what the inference says that only GitHub answers.
//
// The inference reads an asset list the same way whatever published it, so what it writes
// is a GitHub release: a checksum file in it, a signature identity in a workflow of a
// github.com repository, provenance GitHub's verifier checks. The first is the same file
// on the instance and is kept by name; the rest are about somebody else's repository, and
// a definition carrying them is one aqua refuses.
func heldToGitLab(pkgInfo *registry.PackageInfo) {
	heldChecksum(pkgInfo.Checksum)
	pkgInfo.Cosign = nil
	pkgInfo.SLSAProvenance = nil
	pkgInfo.GitHubArtifactAttestations = nil
	// Every version_override says the same things, and more than one release read is
	// what writes them.
	for _, vo := range pkgInfo.VersionOverrides {
		heldChecksum(vo.Checksum)
		vo.Cosign = nil
		vo.SLSAProvenance = nil
		vo.GitHubArtifactAttestations = nil
	}
}

// heldChecksum keeps the checksum file, as the release's own.
//
// It is the same file on the instance, found by the same name, so only what the inference
// calls the release it is in changes. What it says about who signed it doesn't survive: the
// identity is a workflow of a github.com repository.
func heldChecksum(c *registry.Checksum) {
	if c == nil {
		return
	}
	if c.Type == pkgTypeGitHubRelease {
		c.Type = pkgTypeGitLabRelease
	}
	c.Cosign = nil
	c.GitHubArtifactAttestations = nil
}
