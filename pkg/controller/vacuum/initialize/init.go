package initialize

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var errReadPackages = errors.New("some packages couldn't be read from their registries")

func (c *Controller) Init(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	for _, cfgFilePath := range c.configFinder.Finds(param.CWD, param.ConfigFilePath) {
		if err := c.create(ctx, logger, cfgFilePath, param); err != nil {
			return err
		}
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := c.create(ctx, logger, cfgFilePath, param); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) create(ctx context.Context, logger *slog.Logger, cfgFilePath string, param *config.Param) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	lf, err := lockfile.ReadFile(filepath.Join(filepath.Dir(cfgFilePath), lockfile.FileName))
	if err != nil {
		return err //nolint:wrapcheck
	}
	if lf != nil {
		return c.createLocked(ctx, logger, cfg, cfgFilePath, lf)
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgs, _ := config.ListPackages(logger, cfg, c.runtime, registryContents)
	now := time.Now()
	for _, pkg := range pkgs {
		c.createTimestamp(logger, pkg, now)
	}
	return nil
}

// createLocked records the packages of a configuration with a lock file, at the paths
// install puts them in.
//
// Install reads a package's definition through the lock file: the standard registry's from
// aqua-registry-g2, any other registry's at the commit the lock file recorded. The path
// is read from the same definition, since one read from the registry at aqua.yaml's ref
// can name another directory. Only the registries the packages come from are read, so an
// unreadable registry fails only its own packages.
//
// A package without an entry for this environment isn't installed here: install passes
// over a package with no build for it and refuses one the lock file doesn't have. A
// package whose definition can't be read is logged, and the others are still recorded
// before the configuration fails, as install does.
func (c *Controller) createLocked(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, lf *lockfile.LockFile) error {
	src := lockresolve.NewSource(c.g2, lockresolve.NewRegistries(c.registryInstaller, cfg, cfgFilePath, lf))
	now := time.Now()
	failed := false
	for _, pkg := range cfg.Packages {
		logger := logger.With(
			"package_name", pkg.Name,
			"package_version", pkg.Version,
			"registry", pkg.Registry,
		)
		entry := lf.Find(pkg.Name, pkg.Version, c.runtime)
		if entry == nil {
			logger.Debug("skip a package the lock file has no entry for this environment")
			continue
		}
		def, err := src.Locked(ctx, logger, pkg, entry)
		if err != nil {
			slogerr.WithError(logger, err).Error("read the package from its registry")
			failed = true
			continue
		}
		c.createTimestamp(logger, &config.Package{Package: pkg, PackageInfo: def.PackageInfo()}, now)
	}
	if failed {
		return errReadPackages
	}
	return nil
}

// createTimestamp records the package as used now, when it is installed.
//
// A best-effort scan of every configured package: a package that can't be stat'd is
// logged and skipped rather than aborting the whole scan.
func (c *Controller) createTimestamp(logger *slog.Logger, pkg *config.Package, now time.Time) {
	pkgPath, err := pkg.PkgPath(c.runtime)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get a package path")
		return
	}
	absPkgPath := filepath.Join(c.rootDir, pkgPath)
	if f, err := osfile.Exists(absPkgPath); err != nil {
		slogerr.WithError(logger, err).Warn("check if the package is installed")
		return
	} else if !f {
		return
	}
	if err := c.vacuum.Create(pkgPath, now); err != nil {
		slogerr.WithError(logger, err).Warn("create a timestamp file")
	}
}
