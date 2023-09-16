package remove

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Remove(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if param.All {
		logE.Info("removing all packages")
		return c.removeAll(param.RootDir)
	}

	if len(param.Args) == 0 {
		return nil
	}

	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}

	cfg := &aqua.Config{}
	if err := c.configReader.Read(cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
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
		return fmt.Errorf("install registries: %w", err)
	}

	return c.removePackages(logE, param, registryContents)
}

func (c *Controller) removePackages(logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) error {
	for _, pkgName := range param.Args {
		logE := logE.WithField("package_name", pkgName)
		pkg, err := findPkg(pkgName, registryContents)
		if err != nil {
			return fmt.Errorf("find a package from registries: %w", logerr.WithFields(err, logrus.Fields{
				"package_name": pkgName,
			}))
		}
		path := pkg.PkgPath()
		if path == "" {
			logE.WithField("package_type", pkg.Type).Warn("this package type can't be removed")
			continue
		}
		pkgPath := filepath.Join(param.RootDir, "pkgs", path)
		logE.Info("removing a package")
		if err := c.fs.RemoveAll(pkgPath); err != nil {
			return fmt.Errorf("remove a package: %w", logerr.WithFields(err, logrus.Fields{
				"package_name": pkgName,
			}))
		}
	}
	return nil
}

func parsePkgName(pkgName string) (string, string) {
	registryName, pkgName, ok := strings.Cut(pkgName, ",")
	if ok {
		return registryName, pkgName
	}
	return "standard", registryName
}

func findPkg(pkgName string, registryContents map[string]*registry.Config) (*registry.PackageInfo, error) {
	registryName, pkgName := parsePkgName(pkgName)
	rgCfg, ok := registryContents[registryName]
	if !ok {
		return nil, errors.New("unknown registry")
	}
	for _, pkg := range rgCfg.PackageInfos {
		if pkgName != pkg.GetName() {
			continue
		}
		return pkg, nil
	}
	return nil, errors.New("unknown package")
}

func (c *Controller) removeAll(rootDir string) error {
	if err := c.fs.RemoveAll(filepath.Join(rootDir, "pkgs")); err != nil {
		return fmt.Errorf("remove all packages $AQUA_ROOT_DIR/pkgs: %w", err)
	}
	return nil
}
