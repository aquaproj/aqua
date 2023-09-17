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
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Remove(ctx context.Context, logE *logrus.Entry, param *config.Param) error { //nolint:cyclop
	if param.All {
		logE.Info("removing all packages")
		return c.removeAll(param.RootDir)
	}

	if len(param.Args) == 0 && !param.Insert {
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

func (c *Controller) removePackagesInteractively(logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) error {
	var pkgs []*fuzzyfinder.Package
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			pkgs = append(pkgs, &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := c.fuzzyFinder.Find(pkgs)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("find the package: %w", err)
	}
	for _, idx := range idxes {
		pkg := pkgs[idx]
		pkgName := pkg.PackageInfo.GetName()
		logE := logE.WithField("package_name", pkgName)
		if err := c.removePackage(logE, param.RootDir, pkg.PackageInfo); err != nil {
			return fmt.Errorf("remove a package: %w", logerr.WithFields(err, logrus.Fields{
				"package_name": pkgName,
			}))
		}
	}
	return nil
}

func (c *Controller) removePackages(logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) error {
	if param.Insert {
		return c.removePackagesInteractively(logE, param, registryContents)
	}

	for _, pkgName := range param.Args {
		logE := logE.WithField("package_name", pkgName)
		pkg, err := findPkg(pkgName, registryContents)
		if err != nil {
			return fmt.Errorf("find a package from registries: %w", logerr.WithFields(err, logrus.Fields{
				"package_name": pkgName,
			}))
		}
		if err := c.removePackage(logE, param.RootDir, pkg); err != nil {
			return fmt.Errorf("remove a package: %w", logerr.WithFields(err, logrus.Fields{
				"package_name": pkgName,
			}))
		}
	}
	return nil
}

func (c *Controller) removePackage(logE *logrus.Entry, rootDir string, pkg *registry.PackageInfo) error {
	path := pkg.PkgPath()
	if path == "" {
		logE.WithField("package_type", pkg.Type).Warn("this package type can't be removed")
		return nil
	}
	pkgPath := filepath.Join(rootDir, "pkgs", path)
	logE.Info("removing a package")
	if err := c.fs.RemoveAll(pkgPath); err != nil {
		return fmt.Errorf("remove directories: %w", err)
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
