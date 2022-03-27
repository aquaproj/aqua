package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/aquaproj/aqua/pkg/validate"
)

type Controller struct {
	packageInstaller  installpackage.Installer
	rootDir           string
	configFinder      finder.ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
}

func New(rootDir config.RootDir, configFinder finder.ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer) *Controller {
	return &Controller{
		rootDir:           string(rootDir),
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
	}
}

func (ctrl *Controller) Install(ctx context.Context, param *config.Param) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	rootBin := filepath.Join(ctrl.rootDir, "bin")

	if err := util.MkdirAll(rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	if err := ctrl.packageInstaller.InstallProxy(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	for _, cfgFilePath := range ctrl.configFinder.Finds(wd, param.ConfigFilePath) {
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, rootBin, param)
}

func (ctrl *Controller) installAll(ctx context.Context, rootBin string, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range ctrl.configFinder.GetGlobalConfigFilePaths() {
		if _, err := os.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, rootBin, cfgFilePath string, param *config.Param) error {
	cfg := &config.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	if err := validate.Config(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return ctrl.packageInstaller.InstallPackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest) //nolint:wrapcheck
}
