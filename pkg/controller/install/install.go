package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	"github.com/clivm/clivm/pkg/config/aqua"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/clivm/clivm/pkg/installpackage"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	dirPermission os.FileMode = 0o775
)

type Controller struct {
	packageInstaller  installpackage.Installer
	rootDir           string
	configFinder      finder.ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
	fs                afero.Fs
	runtime           *runtime.Runtime
}

func New(param *config.Param, configFinder finder.ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, pkgInstaller installpackage.Installer, fs afero.Fs, rt *runtime.Runtime) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		packageInstaller:  pkgInstaller,
		fs:                fs,
		runtime:           rt,
	}
}

func (ctrl *Controller) Install(ctx context.Context, param *config.Param, logE *logrus.Entry) error {
	rootBin := filepath.Join(ctrl.rootDir, "bin")

	if err := ctrl.fs.MkdirAll(rootBin, dirPermission); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}
	if ctrl.runtime.GOOS == "windows" {
		if err := ctrl.fs.MkdirAll(filepath.Join(ctrl.rootDir, "bat"), dirPermission); err != nil {
			return fmt.Errorf("create the directory: %w", err)
		}
	}

	if err := ctrl.packageInstaller.InstallProxy(ctx, logE); err != nil {
		return err //nolint:wrapcheck
	}

	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param, logE); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, rootBin, param, logE)
}

func (ctrl *Controller) installAll(ctx context.Context, rootBin string, param *config.Param, logE *logrus.Entry) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, rootBin, cfgFilePath, param, logE); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, rootBin, cfgFilePath string, param *config.Param, logE *logrus.Entry) error {
	cfg := &clivm.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return ctrl.packageInstaller.InstallPackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest, logE) //nolint:wrapcheck
}
