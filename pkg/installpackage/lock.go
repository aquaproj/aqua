package installpackage

import (
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Target is a package to install, together with the checksum its download must match
// when one is already known.
//
// The checksum travels beside the package rather than inside it because it belongs to
// one downloaded artifact, whereas config.Package describes the package itself. It
// fills ParamInstallPackage.Checksum, the same slot a checksum read from
// aqua-checksums.json fills; a package resolved from a registry leaves it empty and
// the installer looks the checksum up as it always has.
type Target struct {
	Pkg      *config.Package
	Checksum *checksum.Checksum
}

// SplitByLockFile divides the packages in cfg into those the lock file already
// describes and those it doesn't.
//
// The point of the split is what the caller can skip: a run whose packages are all
// locked needs no registry at all, so nothing is downloaded and nothing is evaluated.
// Packages the lock file doesn't cover fall back to the registry, which is what keeps
// a configuration that predates the lock file working.
func SplitByLockFile(logger *slog.Logger, lf *lockfile.LockFile, cfg *aqua.Config, rt *runtime.Runtime) ([]*Target, []*aqua.Package) {
	locked := make([]*Target, 0, len(cfg.Packages))
	rest := make([]*aqua.Package, 0, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		if pkg.Name == "" || pkg.Version == "" {
			// ListPackages reports these, so leaving them in rest keeps one place
			// where a broken entry is described.
			rest = append(rest, pkg)
			continue
		}
		entry := lf.Find(pkg.Name, pkg.Version, rt)
		if entry == nil {
			rest = append(rest, pkg)
			continue
		}
		logger := logger.With("package_name", pkg.Name, "package_version", pkg.Version)
		t, err := newTarget(pkg, entry, cfg.Registries[pkg.Registry], rt)
		if err != nil {
			// The lock file is wrong about this package rather than silent about
			// it, so falling back to the registry would hide the problem behind a
			// result that looks fine.
			slogerr.WithError(logger, err).Warn("ignore a lock file entry")
			rest = append(rest, pkg)
			continue
		}
		logger.Debug("install the package from the lock file")
		locked = append(locked, t)
	}
	return locked, rest
}

func newTarget(pkg *aqua.Package, entry *lockfile.Package, rgst *aqua.Registry, rt *runtime.Runtime) (*Target, error) {
	if entry.Checksum == "" && entry.NeedsChecksum() {
		return nil, errNoChecksumInLockFile
	}
	p := &config.Package{
		Package:     pkg,
		PackageInfo: entry.PackageInfo(),
		Registry:    rgst,
	}
	t := &Target{Pkg: p}
	if entry.Checksum == "" {
		return t, nil
	}
	// The ID is what aqua-checksums.json keys an entry by. Nothing reads it here,
	// since the checksum is already in hand, but computing it keeps a verification
	// failure naming the same thing it would on the registry path.
	id, err := p.ChecksumID(rt)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	t.Checksum = &checksum.Checksum{
		ID:        id,
		Checksum:  entry.Checksum,
		Algorithm: entry.ChecksumAlgorithm,
	}
	return t, nil
}

// listTargets returns everything to install: the packages the lock file describes,
// which arrive already resolved, followed by the ones that still need a registry.
func (is *Installer) listTargets(logger *slog.Logger, param *ParamInstallPackages) ([]*Target, bool) {
	fromRegistry, failed := config.ListPackages(logger, param.Config, is.runtime, param.Registries)
	targets := make([]*Target, 0, len(fromRegistry)+len(param.LockedPackages))
	for _, pkg := range fromRegistry {
		targets = append(targets, &Target{Pkg: pkg})
	}
	return append(targets, param.LockedPackages...), failed
}

// targetPackages drops the checksums, for the steps that only need the packages.
func targetPackages(targets []*Target) []*config.Package {
	pkgs := make([]*config.Package, len(targets))
	for i, t := range targets {
		pkgs[i] = t.Pkg
	}
	return pkgs
}
