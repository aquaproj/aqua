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
	if err := c.updateCommand(ctx, logE, param); err != nil {
		return err
	}
	if len(param.Args) != 0 {
		return nil
	}
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}
	if err := c.update(ctx, logE, param, cfgFilePath); err != nil {
		return err
	}
	return nil
}

func (c *Controller) updateCommand(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	newVersions := map[string]string{}
	for _, arg := range param.Args {
		findResult, err := c.which.Which(ctx, logE, param, arg)
		if err != nil {
			return fmt.Errorf("find a command: %w", err)
		}
		pkg := findResult.Package
		if newVersion := c.getPackageNewVersion(ctx, logE, param, nil, pkg); newVersion != "" {
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
			newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
		}
		filePath := findResult.ConfigFilePath
		if pkg.Package.FilePath != "" {
			filePath = pkg.Package.FilePath
		}
		if err := c.updateFile(logE, filePath, newVersions); err != nil {
			return fmt.Errorf("update a package: %w", err)
		}
	}
	return nil
}

func (c *Controller) update(ctx context.Context, logE *logrus.Entry, param *config.Param, cfgFilePath string) error { //nolint:cyclop
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	if !param.Insert && !param.OnlyPackage && len(param.Args) == 0 {
		if err := c.updateRegistries(ctx, logE, cfgFilePath, cfg); err != nil {
			return fmt.Errorf("update registries: %w", err)
		}
	}

	if param.OnlyRegistry {
		return nil
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

	registryConfigs, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return c.updatePackages(ctx, logE, param, cfgFilePath, registryConfigs)
}
