package initialize

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

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
		pkgPath, err := pkg.PkgPath(c.runtime)
		if err != nil {
			slogerr.WithError(logger, err).Warn("get a package path")
			continue
		}
		absPkgPath := filepath.Join(c.rootDir, pkgPath)
		// A best-effort scan of every configured package: a package that can't
		// be stat'd is logged and skipped rather than aborting the whole scan.
		if f, err := osfile.Exists(absPkgPath); err != nil {
			slogerr.WithError(logger, err).Warn("check if the package is installed")
			continue
		} else if !f {
			continue
		}
		if err := c.vacuum.Create(pkgPath, now); err != nil {
			slogerr.WithError(logger, err).Warn("create a timestamp file")
		}
	}
	return nil
}
