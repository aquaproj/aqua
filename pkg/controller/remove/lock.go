package remove

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// lockedTarget is a package of aqua.yaml with the lock file entry it is resolved through.
type lockedTarget struct {
	pkg   *aqua.Package
	entry *lockfile.Entry
}

// removeLockedPackages removes the packages named on the command line that the lock file
// has, reading only what each of them needs, and returns the names it couldn't find there.
//
// Without a lock file, where a package is defined is only known by reading every registry.
// The lock file says which registry each package was resolved from, and the standard
// registry is read a package and a version at a time from aqua-registry-g2, so a package
// it has is resolved as which resolves one: an unreadable registry the package doesn't
// come from can't fail removing it.
//
// The paths removed are those of the locked version's definition. For most package types
// the path is the repository's directory, every version of it included. Which repository
// another version uses isn't known from one version's definition, so it isn't looked at.
func (c *Controller) removeLockedPackages(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, pkgNames []string) ([]string, error) {
	lf, err := lockfile.ReadFile(filepath.Join(filepath.Dir(cfgFilePath), lockfile.FileName))
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if lf == nil {
		return pkgNames, nil
	}
	src := lockresolve.NewSource(c.g2, lockresolve.NewRegistries(c.registryInstaller, cfg, cfgFilePath, lf))
	var rest []string
	for _, pkgName := range pkgNames {
		targets := lockedTargets(cfg, lf, c.runtime, pkgName)
		if targets == nil {
			// Not declared in aqua.yaml, or not in the lock file: found the way it is
			// without a lock file, among every registry.
			rest = append(rest, pkgName)
			continue
		}
		logger := logger.With("package_name", pkgName)
		for _, t := range targets {
			// A lock file that disagrees with the registry fails here, as it does for
			// which: it is to be brought up to date rather than worked around by
			// reading every registry.
			def, err := src.Locked(ctx, logger, t.pkg, t.entry)
			if err != nil {
				return nil, fmt.Errorf("read the package from its registry: %w", slogerr.With(err,
					"package_name", pkgName,
					"package_version", t.pkg.Version,
				))
			}
			if err := c.removePackage(logger, c.rootDir, def.PackageInfo()); err != nil {
				return nil, fmt.Errorf("remove a package: %w", slogerr.With(err,
					"package_name", pkgName,
				))
			}
		}
	}
	return rest, nil
}

// lockedTargets is the packages of aqua.yaml a package name on the command line names, each
// with the lock file entry to resolve it through, or nil when there is none or one of them
// has no entry.
//
// A name carries no version, so every version aqua.yaml declares the package at is a
// target. The entry for this environment is preferred, and another environment's entry of
// the same version is used when there is none: the definition it resolves to has the same
// paths for every package type whose path doesn't depend on the environment.
func lockedTargets(cfg *aqua.Config, lf *lockfile.LockFile, rt *runtime.Runtime, arg string) []*lockedTarget {
	registryName, name := parsePkgName(arg)
	var targets []*lockedTarget
	for _, pkg := range cfg.Packages {
		if pkg == nil || pkg.Registry != registryName || pkg.Name != name {
			continue
		}
		entry := lf.Find(pkg.Name, pkg.Version, rt)
		if entry == nil {
			entry = anyEntry(lf, pkg.Name, pkg.Version)
		}
		if entry == nil {
			return nil
		}
		targets = append(targets, &lockedTarget{pkg: pkg, entry: entry})
	}
	return targets
}

// anyEntry is an entry of the package at the version for whichever environment comes first.
func anyEntry(lf *lockfile.LockFile, name, version string) *lockfile.Entry {
	for _, e := range lf.Packages {
		if e.Name == name && e.Version == version {
			return e
		}
	}
	return nil
}
