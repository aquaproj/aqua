package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) Install(ctx context.Context, param *Param) error { //nolint:cyclop,funlen
	cfg := &Config{}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", err)
	}
	param.ConfigFilePath = ctrl.getConfigFilePath(wd, param.ConfigFilePath)
	if param.ConfigFilePath == "" {
		return errConfigFileNotFound
	}
	if err := ctrl.readConfig(param.ConfigFilePath, cfg); err != nil {
		return err
	}
	binDir := filepath.Join(filepath.Dir(param.ConfigFilePath), ".aqua", "bin")

	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	if err := os.MkdirAll(binDir, 0o775); err != nil { //nolint:gomnd
		return fmt.Errorf("create the directory: %w", err)
	}
	inlineRepo := make(map[string]*PackageInfo, len(cfg.InlineRepository))
	for _, pkgInfo := range cfg.InlineRepository {
		inlineRepo[pkgInfo.Name] = pkgInfo
	}

	rootBin := filepath.Join(ctrl.RootDir, "bin")

	if err := os.MkdirAll(rootBin, 0o775); err != nil { //nolint:gomnd
		return fmt.Errorf("create the directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(rootBin, "aqua-proxy")); err != nil {
		if err := ctrl.installProxy(ctx); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(cfg.Packages))
	var flagMutex sync.Mutex
	var failed bool
	maxInstallChan := make(chan struct{}, getMaxParallelism())
	for _, pkg := range cfg.Packages {
		go func(pkg *Package) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			if err := ctrl.installPackage(ctx, inlineRepo, pkg, binDir, param.OnlyLink); err != nil {
				<-maxInstallChan
				logrus.WithFields(logrus.Fields{
					"package_name": pkg.Name,
				}).WithError(err).Error("install the package")
				flagMutex.Lock()
				failed = true
				flagMutex.Unlock()
				return
			}
			<-maxInstallChan
		}(pkg)
	}
	wg.Wait()
	if failed {
		return errInstallFailure
	}
	return nil
}

var errInstallFailure = errors.New("it failed to install some packages")

const defaultMaxParallelism = 5

func getMaxParallelism() int {
	envMaxParallelism := os.Getenv("AQUA_MAX_PARALLELISM")
	if envMaxParallelism == "" {
		return defaultMaxParallelism
	}
	num, err := strconv.Atoi(envMaxParallelism)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"AQUA_MAX_PARALLELISM": envMaxParallelism,
		}).Warn("the environment variable AQUA_MAX_PARALLELISM must be a number")
		return defaultMaxParallelism
	}
	if num <= 0 {
		return defaultMaxParallelism
	}
	return num
}

func (ctrl *Controller) installPackage(ctx context.Context, inlineRepo map[string]*PackageInfo, pkg *Package, binDir string, onlyLink bool) error { //nolint:cyclop
	logE := logrus.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"repository":      pkg.Repository,
	})
	logE.Debug("install the package")
	if pkg.Repository != "inline" {
		return fmt.Errorf("only inline repository is supported (%s)", pkg.Repository)
	}
	pkgInfo, ok := inlineRepo[pkg.Name]
	if !ok {
		return fmt.Errorf("repository isn't found %s", pkg.Name)
	}

	assetName, err := pkgInfo.RenderArtifact(pkg)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	for _, file := range pkgInfo.Files {
		if err := ctrl.createLink(binDir, file); err != nil {
			return err
		}
	}

	if onlyLink {
		logE.WithFields(logrus.Fields{
			"only_link": true,
		}).Debug("skip downloading the package")
		return nil
	}

	pkgPath := getPkgPath(ctrl.RootDir, pkg, pkgInfo, assetName)
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

	for _, file := range pkgInfo.Files {
		warnFileSrc(pkg, pkgInfo, pkgPath, file)
	}

	return nil
}

func warnFileSrc(pkg *Package, pkgInfo *PackageInfo, pkgPath string, file *File) {
	if file.Src.Empty() {
		return
	}
	logE := logrus.WithFields(logrus.Fields{
		"file_name": file.Name,
	})
	fileSrc, err := file.RenderSrc(pkg, pkgInfo)
	if err != nil {
		logE.WithError(err).Warn("render the file.src")
		return
	}
	exePath := filepath.Join(pkgPath, fileSrc)
	logE = logE.WithFields(logrus.Fields{
		"exe_path": exePath,
		"file.src": fileSrc,
	})
	finfo, err := os.Stat(exePath)
	if err != nil {
		logE.WithError(err).Warn("exe_path isn't found. Check file.src is correct")
		return
	}
	if finfo.IsDir() {
		logE.Warn("exe_path is directory. Check file.src is correct")
		return
	}
}

func getPkgPath(aquaRootDir string, pkg *Package, pkgInfo *PackageInfo, assetName string) string {
	return filepath.Join(aquaRootDir, "pkgs", pkgInfo.Type, "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Version, assetName)
}

func (ctrl *Controller) createLink(binDir string, file *File) error {
	linkPath := filepath.Join(binDir, file.Name)
	linkDest := filepath.Join(ctrl.RootDir, "bin", "aqua-proxy")
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
			return ctrl.recreateLink(linkPath, linkDest)
		default:
			return fmt.Errorf("unexpected file mode %s: %s", linkPath, mode.String())
		}
	}
	logrus.WithFields(logrus.Fields{
		"link_file": linkPath,
		"new":       linkDest,
	}).Info("create a symbolic link")
	if err := os.Symlink(linkDest, linkPath); err != nil {
		return fmt.Errorf("create a symbolic link: %w", err)
	}
	return nil
}

func (ctrl *Controller) recreateLink(linkPath, linkDest string) error {
	lnDest, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("read a symbolic link (%s): %w", linkPath, err)
	}
	if linkDest == lnDest {
		return nil
	}
	// recreate link
	logrus.WithFields(logrus.Fields{
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
