package genrgst

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) getCargoPackageInfo(ctx context.Context, logger *slog.Logger, pkgName string) (*registry.PackageInfo, []string) {
	crate := strings.TrimPrefix(pkgName, "crates.io/")
	pkgInfo := &registry.PackageInfo{
		Name:  pkgName,
		Type:  "cargo",
		Crate: crate,
	}
	payload, err := c.cargoClient.GetCrate(ctx, crate)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get a crate metadata by crates.io API")
	}
	if payload != nil && payload.Crate != nil {
		pkgInfo.Description = payload.Crate.Description
		if payload.Crate.Homepage != payload.Crate.Repository {
			pkgInfo.Link = payload.Crate.Homepage
		}
		if repo, ok := strings.CutPrefix(payload.Crate.Repository, "https://github.com/"); ok {
			if repoOwner, repoName, found := strings.Cut(repo, "/"); found {
				pkgInfo.RepoOwner = repoOwner
				pkgInfo.RepoName = repoName
			}
		}
	}
	version, err := c.cargoClient.GetLatestVersion(ctx, crate)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get a latest version by crates.io API")
		return pkgInfo, nil
	}
	return pkgInfo, []string{version}
}
