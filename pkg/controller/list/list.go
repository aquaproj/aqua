package list

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
)

func (c *Controller) List(ctx context.Context, logE *logrus.Entry, param *config.Param) error { //nolint:cyclop
	if param.Installed {
		return c.listInstalled(logE, param)
	}
	cfg := &aqua.Config{}
	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if err := c.configReader.Read(logE, cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	checksums, updateChecksum, err := checksum.Open(
		logE, c.fs, cfgFilePath,
		cfg.ChecksumEnabled(param.EnforceChecksum, param.Checksum))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

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
