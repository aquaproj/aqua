package lockupdate

import (
	"context"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Args are the command's own options, as opposed to the global ones in config.Param.
type Args struct {
	// Force updates entries the lock file already has. Without it they are left
	// alone: a lock file exists to keep an answer, so re-resolving one that is
	// already recorded would defeat it.
	Force bool
	// Packages limits the run to the named packages, each "<name>" or
	// "<name>@<version>". Empty means every package in aqua.yaml.
	Packages []string
}

// Update creates or updates the lock file beside each aqua.yaml that applies.
func (c *Controller) Update(ctx context.Context, logger *slog.Logger, param *config.Param, args *Args) error {
	// No aqua.yaml means nothing to lock. That isn't an error: the command is run
	// from wherever the user happens to be, and a directory outside any aqua
	// project simply has nothing to do.
	for _, cfgFilePath := range c.configFinder.Finds(param.CWD, param.ConfigFilePath) {
		if err := c.updateFile(ctx, logger, cfgFilePath, args); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updateFile(ctx context.Context, logger *slog.Logger, cfgFilePath string, args *Args) error {
	logger = logger.With("config_file_path", cfgFilePath)

	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	lockFilePath := filepath.Join(filepath.Dir(cfgFilePath), lockfile.FileName)
	lf, err := lockfile.ReadFile(lockFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if lf == nil {
		// Creating the file a repository doesn't have yet is exactly what this
		// command is for, so starting from an empty one isn't an assumption.
		lf = lockfile.New()
	}

	updated, failed := c.update(ctx, logger, lf, cfg, cfgFilePath, args)

	// Writing sorts, so an unchanged lock file would still be rewritten. Skipping the
	// write keeps the file's timestamp meaningful and keeps a no-op run out of git.
	if updated {
		if err := lockfile.Write(lockFilePath, lf); err != nil {
			return err //nolint:wrapcheck
		}
	}
	if failed {
		return errUpdateLockFile
	}
	return nil
}

// update brings the packages and the registries in the lock file to what the configuration
// asks for, and says whether anything changed and whether anything couldn't be resolved.
func (c *Controller) update(ctx context.Context, logger *slog.Logger, lf *lockfile.LockFile, cfg *aqua.Config, cfgFilePath string, args *Args) (bool, bool) {
	// A registry file is what the lock file records, and two names in aqua.yaml can
	// point at the same one, so what the run did is turned into files.
	partial := registryFiles(cfg, partialRegistries(cfg, args))
	// --force replaces a registry checksum only when the run resolves every package of
	// the registry. Otherwise the packages it resolves would be read from a new file
	// whose checksum isn't recorded, beside the old one that is.
	rgsts := lockresolve.NewRecordingRegistries(c.registryInstaller, c.commitResolver, cfg, cfgFilePath, lf, func(key *lockfile.Registry) bool {
		return args.Force && !containsRegistryFile(partial, key)
	})
	stale := unrecordedRegistries(lf, cfg, args)
	updated, failedRegistries := c.updatePackages(ctx, logger, lf, rgsts, cfg, stale, args)
	failed := len(failedRegistries) > 0
	failedFiles := registryFiles(cfg, failedRegistries)
	if rgsts.Commit(func(key *lockfile.Registry) bool {
		return !containsRegistryFile(failedFiles, key) && !containsRegistryFile(partial, key)
	}) {
		updated = true
	}
	if pruneRegistries(logger, lf, cfg, args) > 0 {
		updated = true
	}
	return updated, failed
}

// updatePackages brings the lock file to what the configuration asks for, and says
// whether anything changed and the registries of the packages that couldn't be resolved.
func (c *Controller) updatePackages(ctx context.Context, logger *slog.Logger, lf *lockfile.LockFile, rgsts *lockresolve.Registries, cfg *aqua.Config, stale map[string]struct{}, args *Args) (bool, map[string]struct{}) {
	updated := false
	failed := map[string]struct{}{}
	for _, pkg := range cfg.Packages {
		logger := logger.With("package_name", pkg.Name, "package_version", pkg.Version)
		changed, err := c.updatePackage(ctx, logger, lf, rgsts, pkg, stale, args)
		if err != nil {
			// One package that can't be resolved shouldn't cost the rest their
			// entries, so the run keeps going and reports at the end.
			failed[pkg.Registry] = struct{}{}
			slogerr.WithError(logger, err).Error("update the lock file")
			continue
		}
		updated = updated || changed
	}
	if prune(logger, lf, cfg, args) > 0 {
		updated = true
	}
	return updated, failed
}

// lockedRegistries calls fn with each registry file the lock file records a checksum of
// for the packages the run is about, and the package that uses it.
func lockedRegistries(cfg *aqua.Config, args *Args, fn func(pkg *aqua.Package, key *lockfile.Registry)) {
	for _, pkg := range cfg.Packages {
		if pkg == nil || !match(args.Packages, pkg) {
			continue
		}
		rgst := cfg.Registries[pkg.Registry]
		if lockresolve.IsStandard(rgst) {
			continue
		}
		if key := lockresolve.LockFileRegistry(rgst); key != nil {
			fn(pkg, key)
		}
	}
}

// registryFiles is the files the named registries read, for those the lock file records a
// checksum of.
func registryFiles(cfg *aqua.Config, names map[string]struct{}) []*lockfile.Registry {
	files := make([]*lockfile.Registry, 0, len(names))
	for name := range names {
		if key := lockresolve.LockFileRegistry(cfg.Registries[name]); key != nil {
			files = append(files, key)
		}
	}
	return files
}

func containsRegistryFile(files []*lockfile.Registry, key *lockfile.Registry) bool {
	return slices.ContainsFunc(files, key.SameFile)
}

// partialRegistries is the names of the registries some package of aqua.yaml comes from that
// the run doesn't resolve, because the command named its packages.
//
// A checksum recorded for such a registry would sit beside the entries of the packages the
// run left alone, which were resolved from whatever file was there before. The next run
// would then take the registry as recorded and never resolve them again, so the checksum
// is left for a run that resolves every package of the registry.
func partialRegistries(cfg *aqua.Config, args *Args) map[string]struct{} {
	partial := map[string]struct{}{}
	if len(args.Packages) == 0 {
		return partial
	}
	for _, pkg := range cfg.Packages {
		if pkg != nil && !match(args.Packages, pkg) {
			partial[pkg.Registry] = struct{}{}
		}
	}
	return partial
}

// unrecordedRegistries is the names of the registries the run's packages come from whose
// file the lock file has no checksum of.
//
// The entries of a package from such a registry are resolved again even when the lock
// file has them. A registry with no checksum is usually one whose ref aqua.yaml changed,
// and the entries were resolved from the file at the old ref: recording the new file's
// checksum beside them would leave entries that don't match the registry they are now
// read from. It is taken before anything is resolved, since resolving the first package
// records the registry for the rest.
func unrecordedRegistries(lf *lockfile.LockFile, cfg *aqua.Config, args *Args) map[string]struct{} {
	stale := map[string]struct{}{}
	lockedRegistries(cfg, args, func(pkg *aqua.Package, key *lockfile.Registry) {
		// A record without the commit, from a lock file written before the commit was
		// recorded, is recorded again, and so are the entries beside it.
		if r := lf.FindRegistry(key); r == nil || r.Commit == "" {
			stale[pkg.Registry] = struct{}{}
		}
	})
	return stale
}

// pruneRegistries drops the checksums of registry files no package of aqua.yaml comes
// from, and reports how many went. Not when the command named its packages, for the same
// reason as prune.
func pruneRegistries(logger *slog.Logger, lf *lockfile.LockFile, cfg *aqua.Config, args *Args) int {
	if len(args.Packages) > 0 {
		return 0
	}
	var declared []*lockfile.Registry
	lockedRegistries(cfg, args, func(_ *aqua.Package, key *lockfile.Registry) {
		declared = append(declared, key)
	})
	pruned := lf.RetainRegistries(func(r *lockfile.Registry) bool {
		return slices.ContainsFunc(declared, r.SameFile)
	})
	if pruned > 0 {
		logger.Info("dropped the registries aqua.yaml doesn't use", "num_of_registries", pruned)
	}
	return pruned
}

// prune drops the entries aqua.yaml no longer declares, and reports how many went.
//
// What a lock file is for is the packages a configuration asks for, so an entry for one it
// doesn't answers a question nobody puts. A version that was bumped leaves the one before
// it behind, and so does a package that was removed.
//
// Not when the command named its packages. Then aqua.yaml isn't what the run is about, and
// every package it didn't name would go.
//
// A configuration whose import can't be read comes back with fewer packages than it has,
// and dropping on that takes entries for packages that do exist. It is done anyway: the
// file that can't be read is a path that is wrong or a permission that is missing, which
// the next run can't read either, and the lock file is committed -- so what it costs is a
// diff to notice rather than an answer to lose.
func prune(logger *slog.Logger, lf *lockfile.LockFile, cfg *aqua.Config, args *Args) int {
	if len(args.Packages) > 0 {
		return 0
	}
	declared := make(map[string]struct{}, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		if pkg == nil {
			continue
		}
		declared[pkg.Name+"\t"+pkg.Version] = struct{}{}
	}
	pruned := lf.Retain(func(name, version string) bool {
		_, ok := declared[name+"\t"+version]
		return ok
	})
	if pruned > 0 {
		logger.Info("dropped the entries aqua.yaml doesn't ask for", "num_of_entries", pruned)
	}
	return pruned
}

func (c *Controller) updatePackage(ctx context.Context, logger *slog.Logger, lf *lockfile.LockFile, rgsts *lockresolve.Registries, pkg *aqua.Package, stale map[string]struct{}, args *Args) (bool, error) {
	if !match(args.Packages, pkg) {
		return false, nil
	}
	if pkg.Version == "" {
		// version_expr and go_version_file decide the version from the machine the
		// command runs on, so locking one would record an answer that isn't the
		// package's.
		logger.Warn("skip a package whose version isn't written in aqua.yaml")
		return false, nil
	}
	if _, ok := stale[pkg.Registry]; !ok && !args.Force && lf.Has(pkg.Name, pkg.Version) {
		return false, nil
	}

	pkgs, err := c.resolve(ctx, logger, rgsts, pkg)
	if err != nil {
		return false, err
	}

	// Remove before appending rather than replacing in place: --force may resolve a
	// different set of environments than the lock file holds, and a package that
	// dropped one would otherwise keep a stale entry for it.
	lf.Remove(pkg.Name, pkg.Version)
	for _, p := range pkgs {
		lf.Packages = append(lf.Packages, p.Entry())
	}
	logger.Info("update the lock file", "environments", len(pkgs))
	return true, nil
}

// resolve reads the package from where its registry says it lives, with a checksum for
// every environment.
//
// aqua-registry-g2 serves its checksums with the definitions. Any other registry doesn't,
// so they are fetched here, which is the work ar2 does ahead of time for the packages g2
// covers.
func (c *Controller) resolve(ctx context.Context, logger *slog.Logger, rgsts *lockresolve.Registries, pkg *aqua.Package) ([]*lockfile.Package, error) {
	defs, err := lockresolve.Resolve(ctx, logger, c.g2, rgsts, pkg)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	if defs.PkgInfo == nil {
		return defs.Packages, nil
	}
	if err := c.fillChecksums(ctx, logger, defs.Packages, pkg, defs.PkgInfo, defs.SupportedEnvs); err != nil {
		return nil, err
	}
	return defs.Packages, nil
}

// match reports whether pkg is one of the packages named on the command line. Each
// name is "<name>" or "<name>@<version>", the same spelling aqua.yaml uses, so a
// version can be pinned without editing the file first.
func match(names []string, pkg *aqua.Package) bool {
	if len(names) == 0 {
		return true
	}
	for _, name := range names {
		n, version, ok := strings.Cut(name, "@")
		if n != pkg.Name {
			continue
		}
		if !ok || version == pkg.Version {
			return true
		}
	}
	return false
}
