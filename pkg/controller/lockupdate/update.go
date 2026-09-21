package lockupdate

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
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

	rgsts := &registries{
		installer:   c.registryInstaller,
		cfg:         cfg,
		cfgFilePath: cfgFilePath,
	}

	updated := false
	failed := false
	for _, pkg := range cfg.Packages {
		logger := logger.With("package_name", pkg.Name, "package_version", pkg.Version)
		changed, err := c.updatePackage(ctx, logger, lf, rgsts, pkg, args)
		if err != nil {
			// One package that can't be resolved shouldn't cost the rest their
			// entries, so the run keeps going and reports at the end.
			failed = true
			slogerr.WithError(logger, err).Error("update the lock file")
			continue
		}
		updated = updated || changed
	}

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

func (c *Controller) updatePackage(ctx context.Context, logger *slog.Logger, lf *lockfile.LockFile, rgsts *registries, pkg *aqua.Package, args *Args) (bool, error) {
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
	if !args.Force && lf.Has(pkg.Name, pkg.Version) {
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
	lf.Packages = append(lf.Packages, pkgs...)
	logger.Info("update the lock file", "environments", len(pkgs))
	return true, nil
}

// resolve reads the package from where its registry says it lives.
//
// The standard registry is mirrored by aqua-registry-g2, which serves a registry.json
// resolved and checksummed ahead of time. Any other registry has no branch there, so
// its own definition is resolved here instead.
func (c *Controller) resolve(ctx context.Context, logger *slog.Logger, rgsts *registries, pkg *aqua.Package) ([]*lockfile.Package, error) {
	if pkg.Registry == aqua.RegistryTypeStandard {
		return c.g2.Resolve(ctx, logger, pkg.Name, pkg.Version) //nolint:wrapcheck
	}
	return c.resolveFromRegistry(ctx, logger, rgsts, pkg)
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
