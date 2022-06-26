package installpackage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/download"
	"github.com/clivm/clivm/pkg/exec"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/clivm/clivm/pkg/util"
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
	linker            link.Linker
	executor          exec.Executor
}

func isWindows(goos string) bool {
	return goos == "windows"
}

func (inst *installer) InstallPackages(ctx context.Context, cfg *aqua.Config, registries map[string]*registry.Config, binDir string, onlyLink, isTest bool, logE *logrus.Entry) error {
	pkgs, failed := inst.createLinks(cfg, registries, binDir, logE)
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

	if err := pkgInfo.Validate(); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	if pkgInfo.Type == "go_install" && pkg.Package.Version == "latest" {
		return errGoInstallForbidLatest
	}

	assetName, err := pkg.RenderAsset(inst.runtime)
	if err != nil {
		return fmt.Errorf("render the asset name: %w", err)
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}

	if err := inst.downloadWithRetry(ctx, pkg, pkgPath, assetName, logE); err != nil {
		return err
	}

	for _, file := range pkgInfo.GetFiles() {
		if err := inst.checkFileSrc(ctx, pkg, file, logE); err != nil {
			if isTest {
				return fmt.Errorf("check file_src is correct: %w", err)
			}
			logerr.WithError(logE, err).Warn("check file_src is correct")
		}
	}

	return nil
}

func (inst *installer) createLinks(cfg *aqua.Config, registries map[string]*registry.Config, binDir string, logE *logrus.Entry) ([]*config.Package, bool) { //nolint:cyclop,funlen
	pkgs := make([]*config.Package, 0, len(cfg.Packages))
	failed := false
	// registry -> package name -> pkgInfo
	m := make(map[string]map[string]*registry.PackageInfo, len(registries))
	env := inst.runtime.GOOS + "/" + inst.runtime.GOARCH
	for _, pkg := range cfg.Packages {
		if pkg.Name == "" {
			logE.Error("ignore a package because the package name is empty")
			failed = true
			continue
		}
		if pkg.Version == "" {
			logE.Error("ignore a package because the package version is empty")
			failed = true
			continue
		}
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
		pkgInfo, err := getPkgInfoFromRegistries(logE, registries, pkg, m)
		if err != nil {
			logerr.WithError(logE, err).Error("install the package")
			failed = true
			continue
		}

		pkgInfo, err = pkgInfo.Override(pkg.Version, inst.runtime)
		if err != nil {
			logerr.WithError(logE, err).Error("evaluate version constraints")
			failed = true
			continue
		}
		supported, err := pkgInfo.CheckSupported(inst.runtime, env)
		if err != nil {
			logerr.WithError(logE, err).Error("check if the package is supported")
			failed = true
			continue
		}
		if !supported {
			logE.Debug("the package isn't supported on this environment")
			continue
		}
		pkgs = append(pkgs, &config.Package{
			Package:     pkg,
			PackageInfo: pkgInfo,
		})
		for _, file := range pkgInfo.GetFiles() {
			if isWindows(inst.runtime.GOOS) {
				if err := inst.createProxyWindows(file.Name, logE); err != nil {
					logerr.WithError(logE, err).Error("create the proxy file")
					failed = true
				}
				continue
			}
			if err := inst.createLink(filepath.Join(binDir, file.Name), proxyName, logE); err != nil {
				logerr.WithError(logE, err).Error("create the symbolic link")
				failed = true
				continue
			}
		}
	}
	return pkgs, failed
}

func getPkgInfoFromRegistries(logE *logrus.Entry, registries map[string]*registry.Config, pkg *aqua.Package, m map[string]map[string]*registry.PackageInfo) (*registry.PackageInfo, error) {
	pkgInfoMap, ok := m[pkg.Registry]
	if !ok {
		registry, ok := registries[pkg.Registry]
		if !ok {
			return nil, errRegistryNotFound
		}
		pkgInfos := registry.PackageInfos.ToMap(logE)
		m[pkg.Registry] = pkgInfos
		pkgInfoMap = pkgInfos
	}

	pkgInfo, ok := pkgInfoMap[pkg.Name]
	if !ok {
		return nil, errPkgNotFound
	}
	return pkgInfo, nil
}

const maxRetryDownload = 1

func (inst *installer) downloadWithRetry(ctx context.Context, pkg *config.Package, dest, assetName string, logE *logrus.Entry) error {
	logE = logE.WithFields(logrus.Fields{
		"package_name":    pkg.Package.Name,
		"package_version": pkg.Package.Version,
		"registry":        pkg.Package.Registry,
	})
	retryCount := 0
	for {
		logE.Debug("check if the package is already installed")
		finfo, err := inst.fs.Stat(dest)
		if err != nil { //nolint:nestif
			// file doesn't exist
			if err := inst.download(ctx, pkg, dest, assetName, logE); err != nil {
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

func (inst *installer) checkFileSrcGo(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	pkgInfo := pkg.PackageInfo
	exePath := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name)
	if isWindows(inst.runtime.GOOS) {
		exePath += ".exe"
	}
	dir, err := pkg.RenderDir(file, inst.runtime)
	if err != nil {
		return fmt.Errorf("render file dir: %w", err)
	}
	exeDir := filepath.Join(inst.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "src", dir)
	if _, err := inst.fs.Stat(exePath); err == nil {
		return nil
	}
	src := file.Src
	if src == "" {
		src = "."
	}
	logE.WithFields(logrus.Fields{
		"exe_path":     exePath,
		"go_src":       src,
		"go_build_dir": exeDir,
	}).Info("building Go tool")
	if _, err := inst.executor.GoBuild(ctx, exePath, src, exeDir); err != nil {
		return fmt.Errorf("build Go tool: %w", err)
	}
	return nil
}

func (inst *installer) checkFileSrc(ctx context.Context, pkg *config.Package, file *registry.File, logE *logrus.Entry) error {
	fields := logrus.Fields{
		"file_name": file.Name,
	}
	logE = logE.WithFields(fields)

	if pkg.PackageInfo.Type == "go" {
		return inst.checkFileSrcGo(ctx, pkg, file, logE)
	}

	pkgPath, err := pkg.GetPkgPath(inst.rootDir, inst.runtime)
	if err != nil {
		return fmt.Errorf("get the package install path: %w", err)
	}

	fileSrc, err := pkg.RenameFile(logE, inst.fs, pkgPath, file, inst.runtime)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
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
