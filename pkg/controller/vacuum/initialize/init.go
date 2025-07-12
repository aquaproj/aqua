package initialize

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Init(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		err := c.create(ctx, logE, cfgFilePath, param)
		if err != nil {
			return err
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		err := c.create(ctx, logE, cfgFilePath, param)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) create(ctx context.Context, logE *logrus.Entry, cfgFilePath string, param *config.Param) error {
	cfg := &aqua.Config{}

	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	err := c.configReader.Read(logE, cfgFilePath, cfg)
	if err != nil {
		return err //nolint:wrapcheck
	}

	checksums, updateChecksum, err := checksum.Open(
		logE, c.fs, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgs, _ := config.ListPackages(logE, cfg, c.runtime, registryContents)
	now := time.Now()

	for _, pkg := range pkgs {
		pkgPath, err := pkg.PkgPath(c.runtime)
		if err != nil {
			logerr.WithError(logE, err).Warn("get a package path")
			continue
		}

		absPkgPath := filepath.Join(c.rootDir, pkgPath)
		if f, err := afero.Exists(c.fs, absPkgPath); err != nil {
			logerr.WithError(logE, err).Warn("check if the package is installed")
			continue
		} else if !f {
			continue
		}
		err := c.vacuum.Create(pkgPath, now)
		if err != nil {
			logerr.WithError(logE, err).Warn("create a timestamp file")
		}
	}

	return nil
}
