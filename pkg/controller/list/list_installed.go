package list

import (
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) listInstalled(logger *slog.Logger, param *config.Param) error {
	cfgFilePaths := c.configFinder.Finds(param.PWD, param.ConfigFilePath)
	cfgFileMap := map[string]struct{}{}
	for _, cfgFilePath := range cfgFilePaths {
		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

		if err := c.listInstalledByConfig(logger, cfgFilePath); err != nil {
			return slogerr.With(err, //nolint:wrapcheck
				"config_file_path", cfgFilePath,
			)
		}
	}

	if !param.All {
		return nil
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logger := logger.With("config_file_path", cfgFilePath)
		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

		logger.Debug("checking a global configuration file")
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := c.listInstalledByConfig(logger, cfgFilePath); err != nil {
			return slogerr.With(err, //nolint:wrapcheck
				"config_file_path", cfgFilePath,
			)
		}
	}
	return nil
}

func (c *Controller) listInstalledByConfig(logger *slog.Logger, cfgFilePath string) error {
	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		fmt.Fprintln(c.stdout, pkg.Name+"\t"+pkg.Version+"\t"+pkg.Registry)
	}
	return nil
}
