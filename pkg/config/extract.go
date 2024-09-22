package config

import (
	"errors"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
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

		pkgInfo, err = pkgInfo.SetVersion(logE, pkg.Version)
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
		p, err := listPackage(logE, cfg, rt, registries, pkg, m, env)
		if err != nil {
			logerr.WithError(logE, err).Error("ignore a package because the package version is empty")
			failed = true
			continue
		}
		if p == nil {
			continue
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, failed
}

func listPackage(logE *logrus.Entry, cfg *aqua.Config, rt *runtime.Runtime, registries map[string]*registry.Config, pkg *aqua.Package, m map[string]map[string]*registry.PackageInfo, env string) (*Package, error) {
	rgst, ok := cfg.Registries[pkg.Registry]
	if ok {
		if rgst.Ref != "" {
			logE = logE.WithField("registry_ref", rgst.Ref)
		}
	}
	pkgInfo, err := getPkgInfoFromRegistries(logE, registries, pkg, m)
	if err != nil {
		return nil, errors.New("install the package")
	}

	pkgInfo, err = pkgInfo.Override(logE, pkg.Version, rt)
	if err != nil {
		return nil, errors.New("evaluate version constraints")
	}
	supported, err := pkgInfo.CheckSupported(rt, env)
	if err != nil {
		return nil, errors.New("check if the package is supported")
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil, nil //nolint:nilnil
	}
	p := &Package{
		Package:     pkg,
		PackageInfo: pkgInfo,
		Registry:    rgst,
	}
	if err := p.ApplyVars(); err != nil {
		return nil, errors.New("apply the package variable")
	}
	return p, nil
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
