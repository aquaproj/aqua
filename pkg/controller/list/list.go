package list

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	stdout            io.Writer
	configFinder      ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
	fs                afero.Fs
}

func NewController(configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, fs afero.Fs) *Controller {
	return &Controller{
		stdout:            os.Stdout,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		fs:                fs,
	}
}

func (c *Controller) List(ctx context.Context, param *config.Param, logE *logrus.Entry) error { //nolint:cyclop
	cfg := &aqua.Config{}
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	var checksums *checksum.Checksums
	if cfg.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, cfgFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return err //nolint:wrapcheck
	}
	for registryName, registryContent := range registryContents {
		for pkgName := range registryContent.PackageInfos.ToMap(logE) {
			if pkgName == "" {
				logE.Debug("ignore a package because the package name is empty")
				continue
			}
			fmt.Fprintln(c.stdout, registryName+","+pkgName)
		}
	}

	return nil
}
