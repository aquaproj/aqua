package which

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/expr"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/pkgtype"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/validate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type controller struct {
	stdout            io.Writer
	rootDir           string
	configFinder      finder.ConfigFinder
	configReader      reader.ConfigReader
	registryInstaller registry.Installer
	runtime           *runtime.Runtime
	osenv             osenv.OSEnv
	fs                afero.Fs
	linker            link.Linker
	pkgTypes          pkgtype.Packages
}

func (ctrl *controller) Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*Which, error) {
	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		pkg, pkgInfo, file, pkgType, err := ctrl.findExecFile(ctx, cfgFilePath, exeName, logE)
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			return ctrl.whichFile(pkg, pkgInfo, file, pkgType)
		}
	}

	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		pkg, pkgInfo, file, pkgType, err := ctrl.findExecFile(ctx, cfgFilePath, exeName, logE)
		if err != nil {
			return nil, err
		}
		if pkg != nil {
			return ctrl.whichFile(pkg, pkgInfo, file, pkgType)
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

func (ctrl *controller) whichFile(pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File, pkgType pkgtype.Package) (*Which, error) {
	filePath, err := pkgType.GetFilePath(pkg, pkgInfo, file)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	return &Which{
		Package: pkg,
		PkgInfo: pkgInfo,
		File:    file,
		ExePath: filePath,
	}, nil
}

func (ctrl *controller) findExecFile(ctx context.Context, cfgFilePath, exeName string, logE *logrus.Entry) (*config.Package, *config.PackageInfo, *config.File, pkgtype.Package, error) {
	cfg := &config.Config{}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return nil, nil, nil, nil, err //nolint:wrapcheck
	}
	if err := validate.Config(cfg); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("configuration is invalid: %w", err)
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return nil, nil, nil, nil, err //nolint:wrapcheck
	}
	for _, pkg := range cfg.Packages {
		if pkgInfo, file, pkgType := ctrl.findExecFileFromPkg(registryContents, exeName, pkg, logE); pkgInfo != nil {
			return pkg, pkgInfo, file, pkgType, nil
		}
	}
	return nil, nil, nil, nil, nil
}

func (ctrl *controller) findExecFileFromPkg(registries map[string]*config.RegistryContent, exeName string, pkg *config.Package, logE *logrus.Entry) (*config.PackageInfo, *config.File, pkgtype.Package) { //nolint:cyclop
	logE = logE.WithFields(logrus.Fields{
		"registry_name": pkg.Registry,
		"package_name":  pkg.Name,
	})
	registry, ok := registries[pkg.Registry]
	if !ok {
		logE.Warn("registry isn't found")
		return nil, nil, nil
	}

	m, err := registry.PackageInfos.ToMap()
	if err != nil {
		logerr.WithError(logE, err).Warn("registry is invalid")
		return nil, nil, nil
	}

	pkgInfo, ok := m[pkg.Name]
	if !ok {
		logE.Warn("package isn't found")
		return nil, nil, nil
	}

	pkgInfo, err = pkgInfo.Override(pkg.Version, ctrl.runtime)
	if err != nil {
		logerr.WithError(logE, err).Warn("version constraint is invalid")
		return nil, nil, nil
	}

	if pkgInfo.SupportedIf != nil {
		supported, err := expr.EvaluateSupportedIf(pkgInfo.SupportedIf, ctrl.runtime)
		if err != nil {
			logerr.WithError(logE, err).WithField("supported_if", *pkgInfo.SupportedIf).Error("check if the package is supported")
			return nil, nil, nil
		}
		if !supported {
			logE.WithField("supported_if", *pkgInfo.SupportedIf).Debug("the package isn't supported on this environment")
			return nil, nil, nil
		}
	}

	pkgType, ok := ctrl.pkgTypes[pkgInfo.Type]
	if !ok {
		logE.WithField("package_type", pkgInfo.Type).Debug("the package type is unsupported")
		return nil, nil, nil
	}

	for _, file := range pkgType.GetFiles(pkgInfo) {
		if file.Name == exeName {
			return pkgInfo, file, pkgType
		}
	}
	return nil, nil, nil
}
