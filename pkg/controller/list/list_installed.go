package list

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) listInstalled(param *config.Param, logE *logrus.Entry) error {
	cfgFilePaths := c.configFinder.Finds(param.PWD, param.ConfigFilePath)
	cfgFileMap := map[string]struct{}{}
	for _, cfgFilePath := range cfgFilePaths {
		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

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
		if !filepath.IsAbs(cfgFilePath) {
			cfgFilePath = filepath.Join(param.PWD, cfgFilePath)
		}

		if _, ok := cfgFileMap[cfgFilePath]; ok {
			continue
		}
		cfgFileMap[cfgFilePath] = struct{}{}

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
		fmt.Fprintln(c.stdout, pkg.Name+"\t"+pkg.Version+"\t"+pkg.Registry)
	}
	return nil
}
