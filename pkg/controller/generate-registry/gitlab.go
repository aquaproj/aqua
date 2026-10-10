package genrgst

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// gitLabHost is the instance a package name can say a project is on.
//
// Only this one: a package on another GitLab instance is a definition somebody writes,
// and its host is the one thing such a definition has to say. gitlab.com is what a name
// can imply, because a gitlab_release package is on it unless it says otherwise.
const gitLabHost = "gitlab.com"

// getGitLabPackageInfo generates the definition of a project on gitlab.com.
//
// The name says where the project is, the way crates.io/<crate> says a crate: what follows
// the host is the project's path, whose last segment is the project and whose earlier ones
// are the namespace it is in. Nothing else here is GitLab's own -- the asset naming is
// read off the release's asset list by the same inference every other type goes through.
func (c *Controller) getGitLabPackageInfo(ctx context.Context, logger *slog.Logger, pkgName, version string, cfg *Config) (*registry.PackageInfo, []string) {
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

	if description, err := c.gitlab.GetDescription(ctx, gitLabHost, project); err != nil {
		slogerr.WithError(logger, err).Warn("get the project", "project", project)
	} else {
		pkgInfo.Description = description
	}

	release, err := c.gitlab.GetRelease(ctx, gitLabHost, project, version)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get the release", "project", project)
		return pkgInfo, []string{version}
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
	heldToGitLab(pkgInfo)
	return pkgInfo, []string{release.TagName}
}

// heldToGitLab drops what the inference says that only GitHub answers.
//
// The inference reads an asset list the same way whatever published it, so what it writes
// is a GitHub release: a checksum file in it, a signature identity in a workflow of a
// github.com repository, provenance GitHub's verifier checks. The first is the same file
// on the instance and is kept by name; the rest are about somebody else's repository, and
// a definition carrying them is one aqua refuses.
func heldToGitLab(pkgInfo *registry.PackageInfo) {
	if pkgInfo.Checksum != nil {
		if pkgInfo.Checksum.Type == pkgTypeGitHubRelease {
			pkgInfo.Checksum.Type = pkgTypeGitLabRelease
		}
		pkgInfo.Checksum.Cosign = nil
		pkgInfo.Checksum.GitHubArtifactAttestations = nil
	}
	pkgInfo.Cosign = nil
	pkgInfo.SLSAProvenance = nil
	pkgInfo.GitHubArtifactAttestations = nil
}
