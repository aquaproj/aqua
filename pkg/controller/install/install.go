package install

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Install is a main method of "install" command.
// This method is also called by "cp" command.
func (c *Controller) Install(ctx context.Context, logger *slog.Logger, param *config.Param) error {
	if param.Dest == "" {
		// Create a "bin" directory and install aqua-proxy in advance.
		// If param.Dest isn't empty, this means this method is called by "copy" command.
		// If the command is "copy", this block is skipped.
		if err := c.mkBinDir(); err != nil {
			return err
		}
		if err := c.packageInstaller.InstallProxy(ctx, logger); err != nil {
			return fmt.Errorf("install aqua-proxy: %w", err)
		}
	}

	policyCfgs, err := c.policyReader.Read(param.PolicyConfigFilePaths)
	if err != nil {
		return fmt.Errorf("read policy files: %w", err)
	}

	globalPolicyPaths := make(map[string]struct{}, len(param.PolicyConfigFilePaths))
	for _, p := range param.PolicyConfigFilePaths {
		globalPolicyPaths[p] = struct{}{}
	}

	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		policyCfgs, err := c.policyReader.Append(logger, cfgFilePath, policyCfgs, globalPolicyPaths)
		if err != nil {
			return fmt.Errorf("append policy configs: %w", slogerr.With(err,
				"config_file_path", cfgFilePath,
			))
		}
		if err := c.install(ctx, logger, cfgFilePath, policyCfgs, param); err != nil {
			return fmt.Errorf("install packages: %w", slogerr.With(err,
				"config_file_path", cfgFilePath,
			))
		}
	}

	return c.installAll(ctx, logger, param, policyCfgs, globalPolicyPaths)
}

func (c *Controller) mkBinDir() error {
	if err := osfile.MkdirAll(c.fs, filepath.Join(c.rootDir, "bin")); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}
	if c.runtime.IsWindows() {
		if err := c.fs.RemoveAll(filepath.Join(c.rootDir, "bat")); err != nil {
			return fmt.Errorf("remove the bat directory: %w", err)
		}
	}
	return nil
}

func (c *Controller) installAll(ctx context.Context, logger *slog.Logger, param *config.Param, policyConfigs []*policy.Config, globalPolicyPaths map[string]struct{}) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		policyConfigs, err := c.policyReader.Append(logger, cfgFilePath, policyConfigs, globalPolicyPaths)
		if err != nil {
			return fmt.Errorf("append policy configs: %w", slogerr.With(err,
				"config_file_path", cfgFilePath,
			))
		}
		if err := c.install(ctx, logger, cfgFilePath, policyConfigs, param); err != nil {
			return fmt.Errorf("install packages: %w", slogerr.With(err,
				"config_file_path", cfgFilePath,
			))
		}
	}
	return nil
}

func (c *Controller) install(ctx context.Context, logger *slog.Logger, cfgFilePath string, policyConfigs []*policy.Config, param *config.Param) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate the configuration: %w", err)
	}

	checksums, updateChecksum, err := checksum.Open(
		logger, c.fs, cfgFilePath, param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logger, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return c.packageInstaller.InstallPackages(ctx, logger, &installpackage.ParamInstallPackages{ //nolint:wrapcheck
		Config:          cfg,
		Registries:      registryContents,
		ConfigFilePath:  cfgFilePath,
		SkipLink:        c.skipLink,
		Tags:            c.tags,
		ExcludedTags:    c.excludedTags,
		PolicyConfigs:   policyConfigs,
		Checksums:       checksums,
		RequireChecksum: cfg.RequireChecksum(param.EnforceRequireChecksum, param.RequireChecksum),
		DisablePolicy:   param.DisablePolicy,
	})
}
