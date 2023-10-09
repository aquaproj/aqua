package update

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
)

func (c *Controller) newRegistryVersion(ctx context.Context, logE *logrus.Entry, rgst *aqua.Registry) (string, error) {
	if rgst.Type == "local" {
		return "", nil
	}

	logE.Debug("getting the latest release")
	release, _, err := c.github.GetLatestRelease(ctx, rgst.RepoOwner, rgst.RepoName)
	if err != nil {
		return "", fmt.Errorf("get the latest release by GitHub API: %w", err)
	}
	// TODO Get the latest tag if the latest release can't be got.
	return release.GetTagName(), nil
}

// func (c *Controller) updateRegistry(ctx context.Context, logE *logrus.Entry, registryName string, rgst *aqua.Registry) error {
// 	if rgst.Type == "local" {
// 		return nil
// 	}
// 	currentVersion, err := version.NewVersion(rgst.Ref)
// 	if err != nil {
// 		logE.WithError(err).Debug("parse a registry version as semver")
// 		return nil
// 	}
// 	// list versions
// 	releases, _, err := c.github.ListReleases(ctx, rgst.RepoOwner, rgst.RepoName, nil)
// 	if err != nil {
// 		return fmt.Errorf("list releases of the registry: %w", err)
// 	}
// 	for _, release := range releases {
// 		release := release
// 		if release.GetPrerelease() {
// 			logE.Debug("ignore a prerelease")
// 			continue
// 		}
// 		v, err := version.NewVersion(release.GetTagName())
// 		if err != nil {
// 			logE.WithError(err).Debug("parse a release tag as semver")
// 			continue
// 		}
// 		if v.LessThanOrEqual(currentVersion) {
// 			break
// 		}
// 	}
// 	// compare versions
// 	// extract a version
// 	// update ast
// 	return nil
// }
