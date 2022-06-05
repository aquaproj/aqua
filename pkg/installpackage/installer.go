package installpackage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/expr"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const proxyName = "aqua-proxy"

type installer struct {
	rootDir        string
	maxParallelism int
	runtime        *runtime.Runtime
	fs             afero.Fs
	linker         link.Linker
	executor       exec.Executor
	installers     config.PackageTypes
}

func (inst *installer) InstallPackages(ctx context.Context, cfg *aqua.Config, registries map[string]*registry.Config, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error {
	pkgs, failed, err := inst.createLinks(cfg, registries, binDir, logE)
	if err != nil {
		return err
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
				"package_name":    pkg.Package.Name,
				"package_version": pkg.Package.Version,
				"registry":        pkg.Package.Registry,
			})
			if err := inst.InstallPackage(ctx, pkg, isTest, logE); err != nil {
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

func (inst *installer) InstallPackage(ctx context.Context, pkg *config.Package, isTest bool, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	logE.Debug("install the package")
	if err := inst.downloadWithRetry(ctx, pkg, logE); err != nil {
		return err
	}

	for _, file := range pkg.Type.GetFiles(pkgInfo) {
		if err := inst.checkFileSrc(pkg, file, logE); err != nil {
			if isTest {
				return fmt.Errorf("check file_src is correct: %w", err)
			}
			logerr.WithError(logE, err).Warn("check file_src is correct")
		}
	}

	return nil
}

func (inst *installer) createLinks(cfg *aqua.Config, registries map[string]*registry.Config, binDir string, logE *logrus.Entry) ([]*config.Package, bool, error) { //nolint:cyclop
	pkgs := make([]*config.Package, 0, len(cfg.Packages))
	failed := false
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
			return nil, false, fmt.Errorf("evaluate version constraints: %w", err)
		}
		if pkgInfo.SupportedIf != nil {
			supported, err := expr.EvaluateSupportedIf(pkgInfo.SupportedIf, inst.runtime)
			if err != nil {
				logerr.WithError(logE, err).WithField("supported_if", *pkgInfo.SupportedIf).Error("check if the package is supported")
				continue
			}
			if !supported {
				logE.WithField("supported_if", *pkgInfo.SupportedIf).Debug("the package isn't supported on this environment")
				continue
			}
		}
		pkgType, ok := inst.installers[pkgInfo.Type]
		if !ok {
			logE.WithField("package_type", pkgInfo.Type).Error("unsupported package type")
			failed = true
			continue
		}
		pkgs = append(pkgs, &config.Package{
			Package:     pkg,
			PackageInfo: pkgInfo,
			Type:        pkgType,
		})
		for _, file := range pkgInfo.GetFiles() {
			if err := inst.createLink(filepath.Join(binDir, file.Name), proxyName, logE); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}
	return pkgs, failed, nil
}

func getPkgInfoFromRegistries(registries map[string]*registry.Config, pkg *aqua.Package) (*registry.PackageInfo, error) {
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

func (inst *installer) downloadWithRetry(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	retryCount := 0
	for {
		logE.Debug("check if the package is already installed")
		f, err := pkg.Type.CheckInstalled(pkg)
		if err != nil {
			return fmt.Errorf("check if the package is already installed: %w", err)
		}
		if !f { //nolint:nestif
			// file doesn't exist
			if err := pkg.Type.Install(ctx, pkg, logE); err != nil {
				if strings.Contains(err.Error(), "file already exists") {
					if retryCount >= maxRetryDownload {
						return fmt.Errorf("install a package: %w", err)
					}
					retryCount++
					logerr.WithError(logE, err).WithFields(logrus.Fields{
						"retry_count": retryCount,
					}).Info("retry installing the package")
					continue
				}
				return fmt.Errorf("install a package: %w", err)
			}
			return nil
		}
		return nil
	}
}

func (inst *installer) checkFileSrc(pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	filePath, err := pkg.Type.GetFilePath(pkg, file)
	if err != nil {
		return fmt.Errorf("get file path: %w", err)
	}
	fields := logrus.Fields{
		"file_path": filePath,
	}
	logE = logE.WithFields(fields)
	finfo, err := inst.fs.Stat(filePath)
	if err != nil {
		return fmt.Errorf("check file_src is correct: %w", err)
	}
	logE.Debug("check the permission")
	if mode := finfo.Mode().Perm(); !util.IsOwnerExecutable(mode) {
		logE.Debug("add the permission to execute the command")
		if err := inst.fs.Chmod(filePath, util.AllowOwnerExec(mode)); err != nil {
			return logerr.WithFields(errChmod, fields) //nolint:wrapcheck
		}
	}
	return nil
}
