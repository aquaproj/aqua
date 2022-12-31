package which

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	cfgRegistry "github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	stdout            io.Writer
	rootDir           string
	configFinder      ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller domain.RegistryInstaller
	runtime           *runtime.Runtime
	osenv             osenv.OSEnv
	fs                afero.Fs
	linker            domain.Linker
	cosignInstaller   domain.CosignInstaller
}

func (ctrl *Controller) Which(ctx context.Context, logE *logrus.Entry, param *config.Param, exeName string) (*domain.FindResult, error) {
	if err := ctrl.cosignInstaller.InstallCosign(ctx, logE, cosign.Version); err != nil {
		return nil, fmt.Errorf("install Cosign: %w", err)
	}
	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		findResult, err := ctrl.findExecFile(ctx, logE, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			return findResult, nil
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logE := logE.WithField("config_file_path", cfgFilePath)
		logE.Debug("checking a global configuration file")
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		findResult, err := ctrl.findExecFile(ctx, logE, cfgFilePath, exeName)
		if err != nil {
			return nil, err
		}
		if findResult != nil {
			return findResult, nil
		}
	}

	if exePath := ctrl.lookPath(ctrl.osenv.Getenv("PATH"), exeName); exePath != "" {
		return &domain.FindResult{
			ExePath: exePath,
		}, nil
	}
	return nil, logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
		"exe_name": exeName,
	})
}

func (ctrl *Controller) getExePath(findResult *domain.FindResult) (string, error) {
	pkg := findResult.Package
	pkgInfo := pkg.PackageInfo
	file := findResult.File
	if pkg.Package.Version == "" {
		return "", errVersionIsRequired
	}
	if pkg.PackageInfo.Type == "go" {
		return filepath.Join(ctrl.rootDir, "pkgs", pkgInfo.GetType(), "github.com", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version, "bin", file.Name), nil
	}
	fileSrc, err := pkg.GetFileSrc(file, ctrl.runtime)
	if err != nil {
		return "", fmt.Errorf("get file_src: %w", err)
	}
	pkgPath, err := pkg.GetPkgPath(ctrl.rootDir, ctrl.runtime)
	if err != nil {
		return "", fmt.Errorf("get pkg install path: %w", err)
	}
	return filepath.Join(pkgPath, fileSrc), nil
}

func (ctrl *Controller) findExecFile(ctx context.Context, logE *logrus.Entry, cfgFilePath, exeName string) (*domain.FindResult, error) {
	cfg := &aqua.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, logE, cfg, cfgFilePath)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		if findResult := ctrl.findExecFileFromPkg(registryContents, exeName, pkg, logE); findResult != nil {
			findResult.Config = cfg
			findResult.ConfigFilePath = cfgFilePath
			findResult.Package.Registry = cfg.Registries[pkg.Registry]
			return findResult, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (ctrl *Controller) findExecFileFromPkg(registries map[string]*cfgRegistry.Config, exeName string, pkg *aqua.Package, logE *logrus.Entry) *domain.FindResult { //nolint:cyclop
	if pkg.Registry == "" || pkg.Name == "" {
		logE.Debug("ignore a package because the package name or package registry name is empty")
		return nil
	}
	logE = logE.WithFields(logrus.Fields{
		"registry_name": pkg.Registry,
		"package_name":  pkg.Name,
	})
	registry, ok := registries[pkg.Registry]
	if !ok {
		logE.Warn("registry isn't found")
		return nil
	}

	m := registry.PackageInfos.ToMap(logE)

	pkgInfo, ok := m[pkg.Name]
	if !ok {
		logE.Warn("package isn't found")
		return nil
	}

	pkgInfo, err := pkgInfo.Override(pkg.Version, ctrl.runtime)
	if err != nil {
		logerr.WithError(logE, err).Warn("version constraint is invalid")
		return nil
	}

	supported, err := pkgInfo.CheckSupported(ctrl.runtime, ctrl.runtime.GOOS+"/"+ctrl.runtime.GOARCH)
	if err != nil {
		logerr.WithError(logE, err).Error("check if the package is supported")
		return nil
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	for _, file := range pkgInfo.GetFiles() {
		if file.Name == exeName {
			findResult := &domain.FindResult{
				Package: &config.Package{
					Package:     pkg,
					PackageInfo: pkgInfo,
				},
				File: file,
			}
			exePath, err := ctrl.getExePath(findResult)
			if err != nil {
				logE.WithError(err).Error("get the execution file path")
				return nil
			}
			findResult.ExePath = exePath
			return findResult
		}
	}
	return nil
}
