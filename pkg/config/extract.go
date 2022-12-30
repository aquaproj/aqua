package config

import (
	"errors"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	errRegistryNotFound = errors.New("registry isn't found")
	errPkgNotFound      = errors.New("package isn't found in the registry")
)

func ListPackagesNotOverride(logE *logrus.Entry, cfg *aqua.Config, registries map[string]*registry.Config) ([]*Package, bool) {
	pkgs := make([]*Package, 0, len(cfg.Packages))
	failed := false
	// registry -> package name -> pkgInfo
	m := make(map[string]map[string]*registry.PackageInfo, len(registries))
	for _, pkg := range cfg.Packages {
		pkg := pkg
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
			logerr.WithError(logE, err).Error("get the package config from the registry")
			failed = true
			continue
		}

		pkgInfo, err = pkgInfo.SetVersion(pkg.Version)
		if err != nil {
			logerr.WithError(logE, err).Error("evaluate version constraints")
			failed = true
			continue
		}
		pkgs = append(pkgs, &Package{
			Package:     pkg,
			PackageInfo: pkgInfo,
		})
	}
	return pkgs, failed
}

func ListPackages(logE *logrus.Entry, cfg *aqua.Config, rt *runtime.Runtime, registries map[string]*registry.Config) ([]*Package, bool) {
	pkgs := make([]*Package, 0, len(cfg.Packages))
	failed := false
	// registry -> package name -> pkgInfo
	m := make(map[string]map[string]*registry.PackageInfo, len(registries))
	env := rt.Env()
	for _, pkg := range cfg.Packages {
		pkg := pkg
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
		rgst, ok := cfg.Registries[pkg.Registry]
		if ok {
			if rgst.Ref != "" {
				logE = logE.WithField("registry_ref", rgst.Ref)
			}
		}
		pkgInfo, err := getPkgInfoFromRegistries(logE, registries, pkg, m)
		if err != nil {
			logerr.WithError(logE, err).Error("install the package")
			failed = true
			continue
		}

		pkgInfo, err = pkgInfo.Override(pkg.Version, rt)
		if err != nil {
			logerr.WithError(logE, err).Error("evaluate version constraints")
			failed = true
			continue
		}
		supported, err := pkgInfo.CheckSupported(rt, env)
		if err != nil {
			logerr.WithError(logE, err).Error("check if the package is supported")
			failed = true
			continue
		}
		if !supported {
			logE.Debug("the package isn't supported on this environment")
			continue
		}
		pkgs = append(pkgs, &Package{
			Package:     pkg,
			PackageInfo: pkgInfo,
			Registry:    rgst,
		})
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
