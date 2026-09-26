package versiongetter

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// RegistryStandard is the name of the registry aqua-registry-g2 mirrors.
const RegistryStandard = "standard"

type FuzzyGetter struct {
	fuzzyFinder FuzzyFinder
	getter      VersionGetter
	g2          VersionGetter
}

func NewFuzzy(finder FuzzyFinder, getter VersionGetter, g2 *G2VersionGetter) *FuzzyGetter {
	return &FuzzyGetter{
		fuzzyFinder: finder,
		getter:      getter,
		g2:          g2,
	}
}

func (g *FuzzyGetter) Get(ctx context.Context, logger *slog.Logger, registryName string, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) string { //nolint:cyclop
	filters, err := createFilters(pkg)
	if err != nil {
		slogerr.WithError(logger, err).Warn("create filters")
		return ""
	}

	repoName := pkg.RepoOwner + "/" + pkg.RepoName
	logger = logger.With("repository", repoName)
	if useFinder { //nolint:nestif
		logger := logger.With() // Copy logger because g.getter.List has a side effect to change logger
		start := time.Now()
		versions, err := g.list(ctx, logger, registryName, pkg, filters, limit)
		elapsed := time.Since(start)
		if err != nil {
			slogerr.WithError(logger, err).Warn("retrieve package versions")
			return ""
		}
		if versions == nil {
			return ""
		}
		currentVersionIndex := 0
		if currentVersion != "" {
			for i, version := range versions {
				if version.Item == currentVersion {
					version.Item += " (*)"
					currentVersionIndex = i
					break
				}
			}
		}
		idx, err := g.fuzzyFinder.Find(versions, true)
		logger.Debug("retrieve package versions in " + elapsed.String()) // finder's output will overwrite log, so log after it
		if err != nil {
			return ""
		}
		if idx == currentVersionIndex {
			return strings.TrimSuffix(versions[idx].Item, " (*)")
		}
		return versions[idx].Item
	}

	start := time.Now()
	version, err := g.version(ctx, logger, registryName, pkg, filters)
	logger.Debug("retrieve package versions in " + time.Since(start).String())
	if err != nil {
		slogerr.WithError(logger, err).Warn("retrieve package versions")
		return ""
	}
	return version
}

// list and version read from wherever the package's versions come from.
//
// A package from the standard registry is offered the versions aqua-registry-g2 has
// generated a registry.json for, because those are the ones "aqua lock update" can
// lock. A package from any other registry has no branch there, so it keeps asking
// upstream and keeps having its registry's version_filter applied.
func (g *FuzzyGetter) list(ctx context.Context, logger *slog.Logger, registryName string, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	if registryName != RegistryStandard {
		return g.getter.List(ctx, logger, pkg, filters, limit) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	versions, err := g.g2.List(ctx, logger, pkg, filters, limit)
	if !notInG2(logger, err) {
		return versions, err //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	return g.getter.List(ctx, logger, pkg, filters, limit) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
}

func (g *FuzzyGetter) version(ctx context.Context, logger *slog.Logger, registryName string, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	if registryName != RegistryStandard {
		return g.getter.Get(ctx, logger, pkg, filters) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	version, err := g.g2.Get(ctx, logger, pkg, filters)
	if !notInG2(logger, err) {
		return version, err //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	return g.getter.Get(ctx, logger, pkg, filters) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
}

// notInG2 reports whether the failure was aqua-registry-g2 having nothing for the
// package, which is a reason to ask upstream rather than a failure.
//
// Narrowing the choice to what g2 holds is there to stop a version being picked that
// "aqua lock update" can't lock. A package with no branch at all can't be locked at
// any version, so there is nothing left for the narrowing to protect, and refusing to
// offer a version would only make the package unusable while g2 is being filled in.
func notInG2(logger *slog.Logger, err error) bool {
	if !errors.Is(err, g2.ErrNoPackageBranch) {
		return false
	}
	logger.Debug("aqua-registry-g2 doesn't hold the package; asking upstream for its versions")
	return true
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}
