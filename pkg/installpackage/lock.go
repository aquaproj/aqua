package installpackage

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
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
	// Locked says the package came from the lock file. Its signatures were verified
	// when the entry was written, and the checksum the entry recorded stands for them,
	// so the install verifies the checksum alone unless the caller asks for more.
	Locked bool
	// Entry is the lock file entry a locked target came from. Until the target is
	// about to be installed, Pkg knows only the package's name and commands, which is
	// all creating links and filtering by name need: the definition to download from is
	// read from the registry then, and only for what is actually installed.
	Entry *lockfile.Entry
}

// LockSource reads the definition of a locked package from its registry. See
// lockresolve.Locked.
type LockSource interface {
	Locked(ctx context.Context, logger *slog.Logger, pkg *aqua.Package, entry *lockfile.Entry) (*lockfile.Package, error)
}

// SplitByLockFile divides the packages in cfg into those to install from the lock
// file and those to resolve through a registry.
//
// A package the lock file describes but not for this machine is in neither: it has no
// build here, which aqua passes over rather than fails on.
//
// Which one a package takes is decided by the lock file as a whole, not per package.
// No lock file means the repository hasn't adopted one, so everything goes to the
// registries exactly as before. A lock file that exists is the authority on what is
// installed: every package must be in it, and one that is missing is reported rather
// than looked up, because silently reading a registry would install something the lock
// file never recorded a checksum for.
//
// Nothing is read from a registry here. A locked target carries the entry's commands,
// which is enough to create links, and the definition is read when the target is
// installed.
func SplitByLockFile(logger *slog.Logger, lf *lockfile.LockFile, cfg *aqua.Config, rt *runtime.Runtime) ([]*Target, []*aqua.Package, error) {
	if lf == nil {
		return nil, cfg.Packages, nil
	}
	locked := make([]*Target, 0, len(cfg.Packages))
	failed := false
	for _, pkg := range cfg.Packages {
		logger := logger.With("package_name", pkg.Name, "package_version", pkg.Version)
		t, err := lockedTarget(lf, cfg, pkg, rt)
		if err != nil {
			failed = true
			slogerr.WithError(logger, err).Error("install the package from the lock file")
			continue
		}
		if t == nil {
			// The package supports no build for this machine. That is a property
			// of the package rather than a problem with the lock file, and aqua
			// has always passed over such a package rather than failing on it.
			logger.Debug("the package isn't supported on this environment")
			continue
		}
		logger.Debug("install the package from the lock file")
		locked = append(locked, t)
	}
	if failed {
		return nil, nil, errLockFile
	}
	return locked, nil, nil
}

func lockedTarget(lf *lockfile.LockFile, cfg *aqua.Config, pkg *aqua.Package, rt *runtime.Runtime) (*Target, error) {
	if pkg.Name == "" {
		return nil, errPkgNameIsEmpty
	}
	if pkg.Version == "" {
		return nil, errPkgVersionIsEmpty
	}
	entry := lf.Find(pkg.Name, pkg.Version, rt)
	if entry == nil {
		// Two different things look the same here, and the lock file tells them
		// apart. Entries for the package but none for this machine means the
		// package has no build for it, which is what supported_envs says and what
		// aqua passes over. No entries at all means the lock file doesn't know the
		// package, which is a file to bring up to date.
		if lf.Has(pkg.Name, pkg.Version) {
			return nil, nil //nolint:nilnil
		}
		return nil, errNotInLockFile
	}
	if err := entry.CheckCommands(); err != nil {
		return nil, err //nolint:wrapcheck
	}
	return &Target{
		Pkg: &config.Package{
			Package:     pkg,
			PackageInfo: CommandsPackageInfo(pkg.Name, entry),
			Registry:    cfg.Registries[pkg.Registry],
		},
		Locked: true,
		Entry:  entry,
	}, nil
}

// CommandsPackageInfo is a package definition holding nothing but the entry's commands.
//
// It stands in for the definition where only the command names matter, such as creating
// links. It names no download, so nothing can be installed from it.
func CommandsPackageInfo(name string, entry *lockfile.Entry) *registry.PackageInfo {
	files := make([]*registry.File, len(entry.Commands))
	for i, cmd := range entry.Commands {
		files[i] = &registry.File{Name: cmd}
	}
	return &registry.PackageInfo{
		Name:  name,
		Files: files,
	}
}

// resolveLocked reads the definition of a locked target from its registry, with the
// checksum the lock file recorded.
func resolveLocked(ctx context.Context, logger *slog.Logger, src LockSource, t *Target, rt *runtime.Runtime) (*Target, error) {
	if src == nil {
		return nil, errNoLockSource
	}
	def, err := src.Locked(ctx, logger, t.Pkg.Package, t.Entry)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return NewTarget(t.Pkg.Package, def, t.Pkg.Registry, rt)
}

// NewTarget is a target to install a locked package from its definition, which carries
// the lock file's checksum.
func NewTarget(pkg *aqua.Package, def *lockfile.Package, rgst *aqua.Registry, rt *runtime.Runtime) (*Target, error) {
	if def.Checksum == "" && def.NeedsChecksum() {
		return nil, errNoChecksumInLockFile
	}
	p := &config.Package{
		Package:     pkg,
		PackageInfo: def.PackageInfo(),
		Registry:    rgst,
	}
	t := &Target{Pkg: p, Locked: true}
	if def.Checksum == "" {
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
		Checksum:  def.Checksum,
		Algorithm: def.ChecksumAlgorithm,
	}
	return t, nil
}

// listTargets returns everything to install: the packages the lock file describes,
// followed by the ones resolved through a registry.
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
