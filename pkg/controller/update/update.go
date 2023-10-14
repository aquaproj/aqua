package update

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/update/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (c *Controller) Update(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}
	if err := c.update(ctx, logE, cfgFilePath); err != nil {
		return err
	}
	return nil
}

func (c *Controller) update(ctx context.Context, logE *logrus.Entry, cfgFilePath string) error { //nolint:funlen,cyclop
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	// config file path -> config
	// TODO read config files
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	newVersions := map[string]string{}
	for _, rgst := range cfg.Registries {
		newVersion, err := c.newRegistryVersion(ctx, logE, rgst)
		if err != nil {
			return err
		}
		if newVersion == "" {
			continue
		}
		newVersions[rgst.Name] = newVersion
	}

	b, err := afero.ReadFile(c.fs, cfgFilePath)
	if err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse configuration file as YAML: %w", err)
	}

	// TODO consider how to update commit hashes
	updated, err := ast.UpdateRegistries(logE, file, newVersions)
	if err != nil {
		return fmt.Errorf("parse a configuration as YAML to update registries: %w", err)
	}

	if updated {
		stat, err := c.fs.Stat(cfgFilePath)
		if err != nil {
			return fmt.Errorf("get configuration file stat: %w", err)
		}
		if err := afero.WriteFile(c.fs, cfgFilePath, []byte(file.String()), stat.Mode()); err != nil {
			return fmt.Errorf("write the configuration file: %w", err)
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
	return c.updatePackages(ctx, logE, cfgFilePath, registryConfigs)
}
