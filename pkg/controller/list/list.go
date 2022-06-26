package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	"github.com/clivm/clivm/pkg/config/aqua"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	stdout            io.Writer
	configFinder      finder.ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
}

func NewController(configFinder finder.ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
	}
}

func (ctrl *Controller) List(ctx context.Context, param *config.Param, logE *logrus.Entry) error {
	cfg := &aqua.Config{}
	cfgFilePath, err := ctrl.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}
	for registryName, registryContent := range registryContents {
		for pkgName := range registryContent.PackageInfos.ToMapWarn(logE) {
			if pkgName == "" {
				logE.Debug("ignore a package because the package name is empty")
				continue
			}
			fmt.Fprintln(ctrl.stdout, registryName+","+pkgName)
		}
	}

	return nil
}
