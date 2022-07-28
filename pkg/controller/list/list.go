package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	stdout            io.Writer
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	registryInstaller domain.RegistryInstaller
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}

func NewController(configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller) *Controller {
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
