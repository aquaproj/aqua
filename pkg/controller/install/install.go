package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/domain"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	dirPermission os.FileMode = 0o775
)

type Controller struct {
	packageInstaller  domain.PackageInstaller
	rootDir           string
	configFinder      ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
	fs                afero.Fs
	runtime           *runtime.Runtime
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func New(param *config.Param, configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, pkgInstaller domain.PackageInstaller, fs afero.Fs, rt *runtime.Runtime) *Controller {
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

func (ctrl *Controller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
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
		if err := ctrl.install(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}

	return ctrl.installAll(ctx, logE, param)
}

func (ctrl *Controller) installAll(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.install(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) install(ctx context.Context, logE *logrus.Entry, cfgFilePath string) error {
	cfg := &aqua.Config{}
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

	return ctrl.packageInstaller.InstallPackages(ctx, cfg, registryContents, logE) //nolint:wrapcheck
}
