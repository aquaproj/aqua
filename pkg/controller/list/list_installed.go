package list

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) listInstalled(param *config.Param, logE *logrus.Entry) error {
	for _, cfgFilePath := range c.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := c.listInstalledByConfig(cfgFilePath); err != nil {
			return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"config_file_path": cfgFilePath,
			})
		}
	}

	if !param.All {
		return nil
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logE := logE.WithField("config_file_path", cfgFilePath)
		logE.Debug("checking a global configuration file")
		if _, err := c.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := c.listInstalledByConfig(cfgFilePath); err != nil {
			return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
				"config_file_path": cfgFilePath,
			})
		}
	}
	return nil
}

func (c *Controller) listInstalledByConfig(cfgFilePath string) error {
	cfg := &aqua.Config{}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		fmt.Fprintln(c.stdout, pkg.Registry+","+pkg.Name+","+pkg.Version)
	}
	return nil
}
