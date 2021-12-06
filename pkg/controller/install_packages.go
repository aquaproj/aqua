package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) installPackages(ctx context.Context, cfg *Config, registries map[string]*RegistryContent, binDir string, onlyLink, isTest bool) error { //nolint:funlen,cyclop
	var failed bool
	for _, pkg := range cfg.Packages {
		logE := ctrl.logE().WithFields(logrus.Fields{
			"package_name":    pkg.Name,
			"package_version": pkg.Version,
			"registry":        pkg.Registry,
		})
		pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
		if err != nil {
			logerr.WithError(logE, err).Error("install the package")
			failed = true
			continue
		}
		pkgInfo, err = pkgInfo.SetVersion(pkg.Version)
		if err != nil {
			return fmt.Errorf("evaluate version constraints: %w", err)
		}
		for _, file := range pkgInfo.GetFiles() {
			if err := ctrl.createLink(filepath.Join(binDir, file.Name), proxyName); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}

	if onlyLink {
		ctrl.logE().WithFields(logrus.Fields{
			"only_link": true,
		}).Debug("skip downloading the package")
		if failed {
			return errInstallFailure
		}
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(cfg.Packages))
	var flagMutex sync.Mutex
	maxInstallChan := make(chan struct{}, getMaxParallelism())

	handleFailure := func() {
		<-maxInstallChan
		flagMutex.Lock()
		failed = true
		flagMutex.Unlock()
	}

	for _, pkg := range cfg.Packages {
		go func(pkg *Package) {
			maxInstallChan <- struct{}{}
			defer func() {
				wg.Done()
				<-maxInstallChan
			}()
			logE := ctrl.logE().WithFields(logrus.Fields{
				"package_name":    pkg.Name,
				"package_version": pkg.Version,
				"registry":        pkg.Registry,
			})
			if pkg.LazyInstall {
				logE.Debug("skip to install the package")
				return
			}
			pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
			if err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}
			pkgInfo, err = pkgInfo.SetVersion(pkg.Version)
			if err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}

			if err := ctrl.installPackage(ctx, pkgInfo, pkg, isTest); err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}
		}(pkg)
	}
	wg.Wait()
	if failed {
		return errInstallFailure
	}
	return nil
}

func getPkgInfoFromRegistries(registries map[string]*RegistryContent, pkg *Package) (*MergedPackageInfo, error) {
	registry, ok := registries[pkg.Registry]
	if !ok {
		return nil, errRegistryNotFound
	}

	pkgInfos, err := registry.PackageInfos.ToMap()
	if err != nil {
		return nil, fmt.Errorf("convert package infos to map: %w", err)
	}

	pkgInfo, ok := pkgInfos[pkg.Name]
	if !ok {
		return nil, errPkgNotFound
	}
	return pkgInfo, nil
}

func (ctrl *Controller) installPackage(ctx context.Context, pkgInfo *MergedPackageInfo, pkg *Package, isTest bool) error {
	logE := ctrl.logE().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Debug("install the package")
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}
	logE.Debug("check if the package is already installed")
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		if err := ctrl.download(ctx, pkg, pkgInfo, pkgPath, assetName); err != nil {
			return err
		}
	} else {
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	for _, file := range pkgInfo.GetFiles() {
		if err := ctrl.checkFileSrc(pkg, pkgInfo, file); err != nil {
			if isTest {
				return fmt.Errorf("check file_src is correct: %w", err)
			}
			logerr.WithError(logE, err).Warn("check file_src is correct")
		}
	}

	return nil
}

func (ctrl *Controller) checkFileSrc(pkg *Package, pkgInfo *MergedPackageInfo, file *File) error {
	fields := logrus.Fields{
		"file_name": file.Name,
	}
	logE := ctrl.logE().WithFields(fields)

	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}

	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}
	exePath := filepath.Join(pkgPath, fileSrc)

	finfo, err := os.Stat(exePath)
	if err != nil {
		return fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(err, fields))
	}
	if finfo.IsDir() {
		return logerr.WithFields(errExePathIsDirectory, fields) //nolint:wrapcheck
	}

	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !isOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := os.Chmod(exePath, allowOwnerExec(mode)); err != nil {
			return logerr.WithFields(errChmod, fields) //nolint:wrapcheck
		}
	}
	return nil
}

const OwnerExecutable os.FileMode = 64

func isOwnerExecutable(mode os.FileMode) bool {
	return mode&OwnerExecutable != 0
}

func allowOwnerExec(mode os.FileMode) os.FileMode {
	return mode | OwnerExecutable
}

func (ctrl *Controller) createLink(linkPath, linkDest string) error {
	if fileInfo, err := os.Lstat(linkPath); err == nil {
		switch mode := fileInfo.Mode(); {
		case mode.IsDir():
			// if file is a directory, raise error
			return fmt.Errorf("%s has already existed and is a directory", linkPath)
		case mode&os.ModeNamedPipe != 0:
			// if file is a pipe, raise error
			return fmt.Errorf("%s has already existed and is a named pipe", linkPath)
		case mode.IsRegular():
			// TODO if file is a regular file, remove it and create a symlink.
			return fmt.Errorf("%s has already existed and is a regular file", linkPath)
		case mode&os.ModeSymlink != 0:
			return recreateLink(linkPath, linkDest)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	ctrl.logE().WithFields(logrus.Fields{
		"link_file": linkPath,
		"new":       linkDest,
	}).Info("create a symbolic link")
	if err := os.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func recreateLink(linkPath, linkDest string) error {
	lnDest, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("read a symbolic link (%s): %w", linkPath, err)
	}
	if linkDest == lnDest {
		return nil
	}
	// recreate link
	log.New().WithFields(logrus.Fields{
		// TODO add version
		"link_file": linkPath,
		"old":       lnDest,
		"new":       linkDest,
	}).Info("recreate a symbolic link")
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("remove a symbolic link (%s): %w", linkPath, err)
	}
	if err := os.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}
