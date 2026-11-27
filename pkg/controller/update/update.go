package update

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/lockresolve"
)

func (c *Controller) Update(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if err := c.updateCommands(ctx, logger, param); err != nil {
		return err
	}
	if len(param.Args) != 0 {
		return nil
	}

	cfgFilePath, err := c.configFinder.Find(param.CWD, param.ConfigFilePath)
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
	if newVersion == "" {
		v, err := c.findCommandVersion(ctx, logger, param, findResult)
		if err != nil {
			return err
		}
		newVersion = v
	}
	if newVersion != "" {
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

// findCommandVersion finds the new version of the package a command was found in.
//
// In a project with a lock file, the definition which found the command is the lock
// file's entry, which install reads and which has none of the fields that decide which
// tags are versions. So the version is looked for the way "aqua update" does without a
// command: from aqua-registry-g2 for the standard registry, and with the registry's whole
// definition for any other.
func (c *Controller) findCommandVersion(ctx context.Context, logger *slog.Logger, param *config.Param, findResult *which.FindResult) (string, error) {
	pkg := findResult.Package
	lf, err := lockfile.ReadFile(filepath.Join(filepath.Dir(findResult.ConfigFilePath), lockfile.FileName))
	if err != nil {
		return "", err //nolint:wrapcheck
	}
	if lf == nil {
		return c.getPackageNewVersion(ctx, logger, param, nil, pkg), nil
	}
	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, findResult.ConfigFilePath, cfg); err != nil {
		return "", fmt.Errorf("read a configuration file: %w", err)
	}
	// Finding the command read the registry at the commit the lock file recorded, so
	// it has a record and no checksums are needed to read one at its ref.
	rgsts := lockresolve.NewUpdateRegistries(c.registryInstaller, cfg, findResult.ConfigFilePath, lf, checksum.New())
	pkgInfo, source, err := versionSource(ctx, logger, rgsts, pkg.Package)
	if err != nil {
		return "", fmt.Errorf("read the definition of the package to update: %w", err)
	}
	version, err := c.fuzzyGetter.Get(ctx, logger, source, pkgInfo, pkg.Package.Version, param.SelectVersion, param.Limit)
	if err != nil {
		return "", fmt.Errorf("find the new version of the package: %w", err)
	}
	return version, nil
}

func (c *Controller) update(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath string) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	// Update packages before registries because if registries are updated before packages the function needs to install new registries then checksums of new registrires aren't added to aqua-checksums.json.

	if !param.OnlyRegistry {
		if err := c.updateAllPackages(ctx, logger, param, cfgFilePath, cfg, checksums); err != nil {
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

// updateAllPackages updates the packages of the configuration, reading the registries the
// way install does when the project has a lock file.
func (c *Controller) updateAllPackages(ctx context.Context, logger *slog.Logger, param *config.Param, cfgFilePath string, cfg *aqua.Config, checksums *checksum.Checksums) error {
	lf, err := lockfile.ReadFile(filepath.Join(filepath.Dir(cfgFilePath), lockfile.FileName))
	if err != nil {
		return err //nolint:wrapcheck
	}
	if lf != nil {
		return c.updatePackagesLocked(ctx, logger, param, cfgFilePath, lockresolve.NewUpdateRegistries(c.registryInstaller, cfg, cfgFilePath, lf, checksums))
	}
	registryConfigs, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}
	return c.updatePackages(ctx, logger, param, cfgFilePath, registryConfigs)
}
