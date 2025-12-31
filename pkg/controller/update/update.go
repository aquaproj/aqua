package update

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func (c *Controller) Update(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if err := c.updateCommands(ctx, logger, param); err != nil {
		return err
	}
	if len(param.Args) != 0 {
		return nil
	}

	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}
	if err := c.update(ctx, logger, param, cfgFilePath); err != nil {
		return err
	}
	return nil
}

func (c *Controller) updateCommands(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	newVersions := map[string]string{}
	for _, arg := range param.Args {
		if err := c.updateCommand(ctx, logger, param, newVersions, arg); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) updateCommand(ctx context.Context, logger *slog.Logger, param *config.Param, newVersions map[string]string, cmd string) error {
	command, newVersion, _ := strings.Cut(cmd, "@")
	findResult, err := c.which.Which(ctx, logger, param, command)
	if err != nil {
		return fmt.Errorf("find a command: %w", err)
	}

	if findResult.Package == nil {
		return errors.New("command not managed by aqua")
	}

	pkg := findResult.Package
	if newVersion != "" {
		newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
		newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
	} else if newVersion := c.getPackageNewVersion(ctx, logger, param, nil, pkg); newVersion != "" {
		newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.PackageInfo.GetName())] = newVersion
		newVersions[fmt.Sprintf("%s,%s", pkg.Package.Registry, pkg.Package.Name)] = newVersion
	}
	filePath := findResult.ConfigFilePath
	if pkg.Package.FilePath != "" {
		filePath = pkg.Package.FilePath
	}
	if err := c.updateFile(logger, filePath, newVersions); err != nil {
		return fmt.Errorf("update a package: %w", err)
	}
	return nil
}

func (c *Controller) update(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath string) error { //nolint:cyclop
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	// Update packages before registries because if registries are updated before packages the function needs to install new registries then checksums of new registrires aren't added to aqua-checksums.json.

	if !param.OnlyRegistry {
		registryConfigs, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := c.updatePackages(ctx, logger, param, cfgFilePath, registryConfigs); err != nil {
			return err
		}
	}

	if param.Insert || param.OnlyPackage || len(param.Args) != 0 {
		return nil
	}

	if err := c.updateRegistries(ctx, logger, cfgFilePath, cfg); err != nil {
		return fmt.Errorf("update registries: %w", err)
	}
	return nil
}
