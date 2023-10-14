package update

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
)

func (c *Controller) Update(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}
	if err := c.update(ctx, logE, param, cfgFilePath); err != nil {
		return err
	}
	return nil
}

func (c *Controller) update(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	if !param.Insert {
		if err := c.updateRegistries(ctx, logE, cfgFilePath, cfg); err != nil {
			return fmt.Errorf("update registries: %w", err)
		}
	}

	var checksums *checksum.Checksums
	if cfg.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}
	// install registries
	registryConfigs, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	// update packages
	return c.updatePackages(ctx, logE, param, cfgFilePath, registryConfigs)
}
