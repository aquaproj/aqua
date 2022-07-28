package which

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	cfgRegistry "github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type controller struct {
	stdout            io.Writer
	rootDir           string
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	registryInstaller domain.RegistryInstaller
	runtime           *runtime.Runtime
	osenv             osenv.OSEnv
	fs                afero.Fs
	linker            link.Linker
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func (ctrl *controller) Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*Which, error) {
	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		which, err := ctrl.findExecFile(ctx, cfgFilePath, exeName, logE)
		if err != nil {
			return nil, err
		}
		if which != nil {
			return which, nil
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		logE := logE.WithField("config_file_path", cfgFilePath)
		logE.Debug("checking a global configuration file")
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		which, err := ctrl.findExecFile(ctx, cfgFilePath, exeName, logE)
		if err != nil {
			return nil, err
		}
		if which != nil {
			return which, nil
		}
	}

	if exePath := ctrl.lookPath(ctrl.osenv.Getenv("PATH"), exeName); exePath != "" {
		return &Which{
			ExePath: exePath,
		}, nil
	}
	return nil, logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
		"exe_name": exeName,
	})
}

func (ctrl *controller) getExePath(which *Which) (string, error) {
	pkg := which.Package
	pkgInfo := pkg.PackageInfo
	file := which.File
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

func (ctrl *controller) findExecFile(ctx context.Context, cfgFilePath, exeName string, logE *logrus.Entry) (*Which, error) {
	cfg := &aqua.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		if which := ctrl.findExecFileFromPkg(registryContents, exeName, pkg, logE); which != nil {
			return which, nil
		}
	}
	return nil, nil //nolint:nilnil
}

func (ctrl *controller) findExecFileFromPkg(registries map[string]*cfgRegistry.Config, exeName string, pkg *aqua.Package, logE *logrus.Entry) *Which { //nolint:cyclop
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
			which := &Which{
				Package: &config.Package{
					Package:     pkg,
					PackageInfo: pkgInfo,
				},
				File: file,
			}
			exePath, err := ctrl.getExePath(which)
			if err != nil {
				logE.WithError(err).Error("get the execution file path")
				return nil
			}
			which.ExePath = exePath
			return which
		}
	}
	return nil
}
