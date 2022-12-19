package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	stdout            io.Writer
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	registryInstaller domain.RegistryInstaller
	cosignInstaller   domain.CosignInstaller
}

func NewController(configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, cosignInstaller domain.CosignInstaller) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		cosignInstaller:   cosignInstaller,
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

	if err := ctrl.cosignInstaller.InstallCosign(ctx, logE, cosign.Version); err != nil {
		return fmt.Errorf("install Cosign: %w", err)
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath)
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
