package versiongetter

import (
	"context"
	"errors"
	"fmt"
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

// Source is where the versions of a package are chosen from.
type Source int

const (
	// SourceUpstream asks where the package's definition says its versions are, a
	// GitHub repository's releases for one. It is how a package of any registry other
	// than the standard one is updated.
	SourceUpstream Source = iota
	// SourceG2OrUpstream chooses from aqua-registry-g2's versions directory, and asks
	// upstream when g2 has nothing for the package. It is how a project without a lock
	// file updates a package of the standard registry, and how aqua generate chooses
	// one.
	SourceG2OrUpstream
	// SourceG2 chooses from aqua-registry-g2's versions.json and nothing else. It is how
	// a project with a lock file updates a package of the standard registry: g2 is what
	// the lock file is resolved from, so a version g2 doesn't have can't be locked, and
	// g2 has no definition to ask upstream with. A package g2 has nothing for is left
	// alone, and a g2 that can't be read is a failure rather than a reason to ask
	// elsewhere.
	SourceG2
)

// SourceOf is the source a project without a lock file chooses from, which the name
// of the registry decides.
func SourceOf(registryName string) Source {
	if registryName == RegistryStandard {
		return SourceG2OrUpstream
	}
	return SourceUpstream
}

type FuzzyGetter struct {
	fuzzyFinder FuzzyFinder
	getter      VersionGetter
	g2          VersionGetter
	g2Locked    VersionGetter
}

func NewFuzzy(finder FuzzyFinder, getter VersionGetter, g2 *G2VersionGetter) *FuzzyGetter {
	return &FuzzyGetter{
		fuzzyFinder: finder,
		getter:      getter,
		g2:          g2,
		g2Locked:    g2.Locked(),
	}
}

// Get returns the version chosen, or "" when none is.
//
// The error is returned only for SourceG2 when aqua-registry-g2 can't be read. Every
// other failure is logged and comes back as "", as it always has: the package is left
// as it is and the run goes on.
func (g *FuzzyGetter) Get(ctx context.Context, logger *slog.Logger, source Source, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) (string, error) {
	if source == SourceG2 {
		return g.getFromG2(ctx, logger, pkg, currentVersion, useFinder, limit)
	}
	return g.get(ctx, logger, source, pkg, currentVersion, useFinder, limit), nil
}

// getFromG2 chooses from aqua-registry-g2 alone. See SourceG2.
func (g *FuzzyGetter) getFromG2(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) (string, error) {
	if useFinder {
		versions, err := g.g2Locked.List(ctx, logger, pkg, nil, limit)
		if skip, err := g2Outcome(logger, err, len(versions) == 0); skip || err != nil {
			return "", err
		}
		return g.find(logger, versions, currentVersion, 0), nil
	}
	version, err := g.g2Locked.Get(ctx, logger, pkg, nil)
	if skip, err := g2Outcome(logger, err, version == ""); skip || err != nil {
		return "", err
	}
	return version, nil
}

// g2Outcome tells apart aqua-registry-g2 holding nothing for the package, which leaves
// it alone (skip), from g2 not being read, which fails it.
func g2Outcome(logger *slog.Logger, err error, empty bool) (bool, error) {
	if err != nil && !errors.Is(err, g2.ErrNoPackageBranch) {
		return false, fmt.Errorf("retrieve package versions from aqua-registry-g2: %w", err)
	}
	if err == nil && !empty {
		return false, nil
	}
	logger.Warn("skip updating the package because aqua-registry-g2 has no version of it")
	return true, nil
}

// find lets the user choose from the versions, marking the current one.
func (g *FuzzyGetter) find(logger *slog.Logger, versions []*fuzzyfinder.Item, currentVersion string, elapsed time.Duration) string {
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
	if elapsed != 0 {
		logger.Debug("retrieve package versions in " + elapsed.String()) // finder's output will overwrite log, so log after it
	}
	if err != nil {
		return ""
	}
	if idx == currentVersionIndex {
		return strings.TrimSuffix(versions[idx].Item, " (*)")
	}
	return versions[idx].Item
}

func (g *FuzzyGetter) get(ctx context.Context, logger *slog.Logger, source Source, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) string {
	filters, err := createFilters(pkg)
	if err != nil {
		slogerr.WithError(logger, err).Warn("create filters")
		return ""
	}

	repoName := pkg.RepoOwner + "/" + pkg.RepoName
	logger = logger.With("repository", repoName)
	if useFinder {
		logger := logger.With() // Copy logger because g.getter.List has a side effect to change logger
		start := time.Now()
		versions, err := g.list(ctx, logger, source, pkg, filters, limit)
		elapsed := time.Since(start)
		if err != nil {
			slogerr.WithError(logger, err).Warn("retrieve package versions")
			return ""
		}
		if versions == nil {
			return ""
		}
		return g.find(logger, versions, currentVersion, elapsed)
	}

	start := time.Now()
	version, err := g.version(ctx, logger, source, pkg, filters)
	logger.Debug("retrieve package versions in " + time.Since(start).String())
	if err != nil {
		slogerr.WithError(logger, err).Warn("retrieve package versions")
		return ""
	}
	return version
}

// list and version read from wherever the package's versions come from, for
// SourceUpstream and SourceG2OrUpstream.
//
// A package from the standard registry is offered the versions aqua-registry-g2 has
// generated a registry.json for, because those are the ones "aqua lock update" can
// lock. A package from any other registry has no branch there, so it keeps asking
// upstream and keeps having its registry's version_filter applied.
func (g *FuzzyGetter) list(ctx context.Context, logger *slog.Logger, source Source, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	if source != SourceG2OrUpstream {
		return g.getter.List(ctx, logger, pkg, filters, limit) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	versions, err := g.g2.List(ctx, logger, pkg, filters, limit)
	if !notInG2(logger, err) && len(versions) > 0 {
		return versions, err //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	return g.getter.List(ctx, logger, pkg, filters, limit) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
}

func (g *FuzzyGetter) version(ctx context.Context, logger *slog.Logger, source Source, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	if source != SourceG2OrUpstream {
		return g.getter.Get(ctx, logger, pkg, filters) //nolint:wrapcheck // the caller reports it as the failure to retrieve versions
	}
	version, err := g.g2.Get(ctx, logger, pkg, filters)
	if !notInG2(logger, err) && version != "" {
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
//
// A branch holding no version is the same situation: the branch is made when the
// package is taken over and the versions are generated afterwards, so it exists empty
// for a while. That is why an empty answer falls back too, rather than only an
// explicit "no branch".
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
