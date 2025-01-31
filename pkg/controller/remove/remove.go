package remove

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
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

	cmds := make([]string, 0, len(param.Args))
	pkgs := make([]string, 0, len(param.Args))
	for _, arg := range param.Args {
		if strings.Contains(arg, "/") {
			pkgs = append(pkgs, arg)
			continue
		}
		cmds = append(cmds, arg)
	}

	if err := c.removeCommands(ctx, logE, param, cmds); err != nil {
		return err
	}

	param.Args = pkgs

	cfgFilePath, err := c.configFinder.Find(param.PWD, param.ConfigFilePath, param.GlobalConfigFilePaths...)
	if err != nil {
		return fmt.Errorf("find a configuration file: %w", err)
	}

	cfg := &aqua.Config{}
	if err := c.configReader.Read(logE, cfgFilePath, cfg); err != nil {
		return fmt.Errorf("read a configuration file: %w", err)
	}

	checksums, updateChecksum, err := checksum.Open(
		logE, c.fs, cfgFilePath,
		param.ChecksumEnabled(cfg))
	if err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	defer updateChecksum()

	registryContents, err := c.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath, checksums)
	if err != nil {
		return fmt.Errorf("install registries: %w", err)
	}

	return c.removePackages(logE, param, registryContents)
}

func (c *Controller) removePackagesInteractively(logE *logrus.Entry, param *config.Param, registryContents map[string]*registry.Config) error {
	var items []*fuzzyfinder.Item
	var pkgs []*fuzzyfinder.Package
	for registryName, registryContent := range registryContents {
		for _, pkg := range registryContent.PackageInfos {
			fp := &fuzzyfinder.Package{
				PackageInfo:  pkg,
				RegistryName: registryName,
			}
			pkgs = append(pkgs, fp)
			items = append(items, &fuzzyfinder.Item{
				Item:    fp.Item(),
				Preview: fuzzyfinder.PreviewPackage(fp),
			})
		}
	}

	// Launch the fuzzy finder
	idxes, err := c.fuzzyFinder.FindMulti(items, true)
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

func (c *Controller) removeCommands(ctx context.Context, logE *logrus.Entry, param *config.Param, cmds []string) error {
	var gErr error
	if c.mode.Link {
		for _, cmd := range cmds {
			logE := logE.WithField("exe_name", cmd)
			logE.Info("removing a link")
			if err := c.fs.Remove(filepath.Join(c.rootDir, "bin", cmd)); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				logerr.WithError(logE, err).Error("remove a link")
				gErr = errors.New("remove links")
			}
		}
	}
	if !c.mode.Package {
		return gErr
	}
	for _, cmd := range cmds {
		logE := logE.WithField("exe_name", cmd)
		if err := c.removeCommand(ctx, logE, param, cmd); err != nil {
			return fmt.Errorf("remove a command: %w", logerr.WithFields(err, logrus.Fields{
				"exe_name": cmd,
			}))
		}
	}
	return gErr
}

func (c *Controller) removeCommand(ctx context.Context, logE *logrus.Entry, param *config.Param, cmd string) error {
	findResult, err := c.which.Which(ctx, logE, param, cmd)
	if err != nil {
		if errors.Is(err, which.ErrCommandIsNotFound) {
			logE.Debug("the command isn't found")
			return nil
		}
		return fmt.Errorf("find a command: %w", err)
	}
	if findResult.Package == nil {
		logE.Debug("no package is found")
		return nil
	}
	logE = logE.WithField("package_name", findResult.Package.Package.Name)
	if err := c.removePackage(logE, param.RootDir, findResult.Package.PackageInfo); err != nil {
		return fmt.Errorf("remove a package: %w", logerr.WithFields(err, logrus.Fields{
			"package_name": findResult.Package.Package.Name,
		}))
	}
	return nil
}

func (c *Controller) removePackage(logE *logrus.Entry, rootDir string, pkg *registry.PackageInfo) error {
	var gErr error
	logE.Info("removing a package")
	if c.mode.Link {
		for _, file := range pkg.GetFiles() {
			if err := c.fs.Remove(filepath.Join(rootDir, "bin", file.Name)); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				logerr.WithError(logE, err).WithField("link", file.Name).Error("remove a link")
				gErr = errors.New("remove links")
			}
		}
	}
	if !c.mode.Package {
		return gErr
	}

	paths := pkg.PkgPaths()
	if len(paths) == 0 {
		logE.WithField("package_type", pkg.Type).Warn("this package can't be removed")
		return gErr
	}
	for path := range paths {
		if err := c.removePath(logE, rootDir, path); err != nil {
			return err
		}
		if err := c.removeMetadataPath(logE, rootDir, path); err != nil {
			return err
		}
	}
	return gErr
}

func (c *Controller) removePath(logE *logrus.Entry, rootDir string, path string) error {
	pkgPath := filepath.Join(rootDir, "pkgs", path)
	arr, err := afero.Glob(c.fs, pkgPath)
	if err != nil {
		return fmt.Errorf("find directories: %w", err)
	}
	for _, p := range arr {
		logE.WithField("removed_path", p).Debug("removing a directory")
		if err := c.fs.RemoveAll(p); err != nil {
			return fmt.Errorf("remove directories: %w", err)
		}
	}
	return nil
}

func (c *Controller) removeMetadataPath(logE *logrus.Entry, rootDir string, path string) error {
	pkgPath := filepath.Join(rootDir, "metadata", "pkgs", path)
	arr, err := afero.Glob(c.fs, pkgPath)
	if err != nil {
		return fmt.Errorf("find directories: %w", err)
	}
	for _, p := range arr {
		logE.WithField("removed_path", p).Debug("removing a directory")
		if err := c.fs.RemoveAll(p); err != nil {
			return fmt.Errorf("remove directories: %w", err)
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
	var gErr error
	if c.mode.Link {
		if err := c.fs.RemoveAll(filepath.Join(rootDir, "bin")); err != nil {
			gErr = fmt.Errorf("remove the bin directory: %w", err)
		}
	}
	if !c.mode.Package {
		return gErr
	}
	if err := c.fs.RemoveAll(filepath.Join(rootDir, "pkgs")); err != nil {
		return fmt.Errorf("remove all packages: %w", err)
	}
	if err := c.fs.RemoveAll(filepath.Join(rootDir, "metadata")); err != nil {
		return fmt.Errorf("remove all package metadata: %w", err)
	}
	return gErr
}
