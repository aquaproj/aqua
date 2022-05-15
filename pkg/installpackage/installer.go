package installpackage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/util"
	constraint "github.com/aquaproj/aqua/pkg/version-constraint"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const proxyName = "aqua-proxy"

type installer struct {
	rootDir           string
	maxParallelism    int
	packageDownloader download.PackageDownloader
	runtime           *runtime.Runtime
	fs                afero.Fs
}

func (inst *installer) InstallPackages(ctx context.Context, cfg *config.Config, registries map[string]*config.RegistryContent, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error { //nolint:funlen,cyclop,gocognit
	var failed bool
	pkgs := make([]*config.Package, 0, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		logE := logE.WithFields(logrus.Fields{
			"package_name":    pkg.Name,
			"package_version": pkg.Version,
			"registry":        pkg.Registry,
		})
		if registry, ok := cfg.Registries[pkg.Registry]; ok {
			if registry.Ref != "" {
				logE = logE.WithField("registry_ref", registry.Ref)
			}
		}
		pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
		if err != nil {
			logerr.WithError(logE, err).Error("install the package")
			failed = true
			continue
		}

		pkgInfo, err = pkgInfo.Override(pkg.Version, inst.runtime)
		if err != nil {
			return fmt.Errorf("evaluate version constraints: %w", err)
		}
		if pkgInfo.SupportedIf != nil {
			supported, err := constraint.EvaluateSupportedIf(pkgInfo.SupportedIf, inst.runtime)
			if err != nil {
				logerr.WithError(logE, err).WithField("supported_if", *pkgInfo.SupportedIf).Error("check if the package is supported")
				continue
			}
			if !supported {
				logE.WithField("supported_if", *pkgInfo.SupportedIf).Debug("the package isn't supported on this environment")
				continue
			}
		}
		pkgs = append(pkgs, pkg)
		for _, file := range pkgInfo.GetFiles() {
			if err := inst.createLink(filepath.Join(binDir, file.Name), proxyName, logE); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}

	if onlyLink {
		logE.WithFields(logrus.Fields{
			"only_link": true,
		}).Debug("skip downloading the package")
		if failed {
			return errInstallFailure
		}
		return nil
	}

	if len(pkgs) == 0 {
		if failed {
			return errInstallFailure
		}
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(pkgs))
	var flagMutex sync.Mutex
	maxInstallChan := make(chan struct{}, inst.maxParallelism)

	handleFailure := func() {
		flagMutex.Lock()
		failed = true
		flagMutex.Unlock()
	}

	for _, pkg := range pkgs {
		go func(pkg *config.Package) {
			defer wg.Done()
			maxInstallChan <- struct{}{}
			defer func() {
				<-maxInstallChan
			}()
			logE := logE.WithFields(logrus.Fields{
				"package_name":    pkg.Name,
				"package_version": pkg.Version,
				"registry":        pkg.Registry,
			})
			if registry, ok := cfg.Registries[pkg.Registry]; ok {
				if registry.Ref != "" {
					logE = logE.WithField("registry_ref", registry.Ref)
				}
			}
			pkgInfo, err := getPkgInfoFromRegistries(registries, pkg)
			if err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}

			pkgInfo, err = pkgInfo.Override(pkg.Version, inst.runtime)
			if err != nil {
				logerr.WithError(logE, err).Error("install the package")
				handleFailure()
				return
			}

			if err := inst.InstallPackage(ctx, pkgInfo, pkg, isTest, logE); err != nil {
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

func getPkgInfoFromRegistries(registries map[string]*config.RegistryContent, pkg *config.Package) (*config.PackageInfo, error) {
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

const maxRetryDownload = 1

func (inst *installer) downloadWithRetry(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, dest, assetName string, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	retryCount := 0
	for {
		logE.Debug("check if the package is already installed")
		finfo, err := inst.fs.Stat(dest)
		if err != nil { //nolint:nestif
			// file doesn't exist
			if err := inst.download(ctx, pkg, pkgInfo, dest, assetName, logE); err != nil {
				if strings.Contains(err.Error(), "file already exists") {
					if retryCount >= maxRetryDownload {
						return err
					}
					retryCount++
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"retry_count": retryCount,
					}).Info("retry installing the package")
					continue
				}
				return err
			}
			return nil
		}
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", dest)
		}
		return nil
	}
}

func (inst *installer) InstallPackage(ctx context.Context, pkgInfo *config.PackageInfo, pkg *config.Package, isTest bool, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})
	logE.Debug("install the package")

	if err := pkgInfo.Validate(); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}
	assetName, err := pkgInfo.RenderAsset(pkg, inst.runtime)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath, err := pkgInfo.GetPkgPath(inst.rootDir, pkg, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}

	if err := inst.downloadWithRetry(ctx, pkg, pkgInfo, pkgPath, assetName, logE); err != nil {
		return err
	}

	for _, file := range pkgInfo.GetFiles() {
		if err := inst.checkFileSrc(pkg, pkgInfo, file, logE); err != nil {
			if isTest {
				return fmt.Errorf("check file_src is correct: %w", err)
			}
			logerr.WithError(logE, err).Warn("check file_src is correct")
		}
	}

	return nil
}

func (inst *installer) checkFileSrc(pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File, logE *logrus.Entry) error {
	fields := logrus.Fields{
		"file_name": file.Name,
	}
	logE = logE.WithFields(fields)

	fileSrc, err := pkgInfo.GetFileSrc(pkg, file, inst.runtime)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}

	pkgPath, err := pkgInfo.GetPkgPath(inst.rootDir, pkg, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}
	exePath := filepath.Join(pkgPath, fileSrc)

	finfo, err := inst.fs.Stat(exePath)
	if err != nil {
		return fmt.Errorf("exe_path isn't found: %w", logerr.WithFields(err, fields))
	}
	if finfo.IsDir() {
		return logerr.WithFields(errExePathIsDirectory, fields) //nolint:wrapcheck
	}

	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !util.IsOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := inst.fs.Chmod(exePath, util.AllowOwnerExec(mode)); err != nil {
			return logerr.WithFields(errChmod, fields) //nolint:wrapcheck
		}
	}
	return nil
}
