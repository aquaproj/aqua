package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/validate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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
}

func New(param *config.Param, configFinder finder.ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, pkgInstaller installpackage.Installer, fs afero.Fs) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		packageInstaller:  pkgInstaller,
		fs:                fs,
	}
}

func (ctrl *Controller) Install(ctx context.Context, param *config.Param, logE *logrus.Entry) error {
	rootBin := filepath.Join(ctrl.rootDir, "bin")

	if err := ctrl.fs.MkdirAll(rootBin, dirPermission); err != nil {
		return fmt.Errorf("create the directory: %w", err)
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
	cfg := &config.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	checksumFile := checksum.GetChecksumFilePathFromConfigFilePath(cfgFilePath)
	if err := ctrl.packageInstaller.ReadChecksumFile(ctrl.fs, checksumFile); err != nil {
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"checksum_file": checksumFile,
		}).Error("read a checksum file")
	}

	if err := validate.Config(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}

	errInstallPackages := ctrl.packageInstaller.InstallPackages(ctx, cfg, registryContents, rootBin, param.OnlyLink, param.IsTest, logE)
	if err := ctrl.packageInstaller.UpdateChecksumFile(ctrl.fs, checksumFile); err != nil {
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"checksum_file": checksumFile,
		}).Error("update a checksum file")
	}
	return errInstallPackages //nolint:wrapcheck
}
