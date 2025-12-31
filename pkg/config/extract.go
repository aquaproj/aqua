package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	// errRegistryNotFound is returned when a referenced registry is not found
	errRegistryNotFound = errors.New("registry isn't found")
	// errPkgNotFound is returned when a package is not found in any registry
	errPkgNotFound = errors.New("package isn't found in the registry")
)

// ListPackagesNotOverride extracts packages from configuration without applying version overrides.
// It validates package definitions and resolves registry information but skips version constraints.
func ListPackagesNotOverride(logger *slog.Logger, cfg *aqua.Config, registries map[string]*registry.Config) ([]*Package, bool) {
	pkgs := make([]*Package, 0, len(cfg.Packages))
	failed := false
	// registry -> package name -> pkgInfo
	m := make(map[string]map[string]*registry.PackageInfo, len(registries))
	for _, pkg := range cfg.Packages {
		if pkg.Name == "" {
			logger.Error("ignore a package because the package name is empty")
			failed = true
			continue
		}
		if pkg.Version == "" {
			logger.Error("ignore a package because the package version is empty")
			failed = true
			continue
		}
		logger := logger.With(
			"package_name", pkg.Name,
			"package_version", pkg.Version,
			"registry", pkg.Registry,
		)
		if registry, ok := cfg.Registries[pkg.Registry]; ok {
			if registry.Ref != "" {
				logger = logger.With("registry_ref", registry.Ref)
			}
		}
		pkgInfo, err := getPkgInfoFromRegistries(logger, registries, pkg, m)
		if err != nil {
			slogerr.WithError(logger, err).Error("get the package config from the registry")
			failed = true
			continue
		}

		pkgInfo, err = pkgInfo.SetVersion(logger, pkg.Version)
		if err != nil {
			slogerr.WithError(logger, err).Error("evaluate version constraints")
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

// ListPackages extracts and validates all packages from configuration.
// It applies version overrides, checks platform support, and processes package variables.
func ListPackages(logger *slog.Logger, cfg *aqua.Config, rt *runtime.Runtime, registries map[string]*registry.Config) ([]*Package, bool) {
	pkgs := make([]*Package, 0, len(cfg.Packages))
	failed := false
	// registry -> package name -> pkgInfo
	m := make(map[string]map[string]*registry.PackageInfo, len(registries))
	env := rt.Env()
	for _, pkg := range cfg.Packages {
		if pkg.Name == "" {
			logger.Error("ignore a package because the package name is empty")
			failed = true
			continue
		}
		if pkg.Version == "" {
			logger.Error("ignore a package because the package version is empty")
			failed = true
			continue
		}
		logger := logger.With(
			"package_name", pkg.Name,
			"package_version", pkg.Version,
			"registry", pkg.Registry,
		)
		p, err := listPackage(logger, cfg, rt, registries, pkg, m, env)
		if err != nil {
			slogerr.WithError(logger, err).Error("ignore a package because the package version is empty")
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

// listPackage processes a single package definition with full validation.
// It applies overrides, checks platform support, and validates package configuration.
func listPackage(logger *slog.Logger, cfg *aqua.Config, rt *runtime.Runtime, registries map[string]*registry.Config, pkg *aqua.Package, m map[string]map[string]*registry.PackageInfo, env string) (*Package, error) {
	rgst, ok := cfg.Registries[pkg.Registry]
	if ok {
		if rgst.Ref != "" {
			logger = logger.With("registry_ref", rgst.Ref)
		}
	}
	pkgInfo, err := getPkgInfoFromRegistries(logger, registries, pkg, m)
	if err != nil {
		return nil, fmt.Errorf("install the package: %w", err)
	}

	pkgInfo, err = pkgInfo.Override(logger, pkg.Version, rt)
	if err != nil {
		return nil, fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(rt, env)
	if err != nil {
		return nil, fmt.Errorf("check if the package is supported: %w", err)
	}
	if !supported {
		logger.Debug("the package isn't supported on this environment")
		return nil, nil //nolint:nilnil
	}
	p := &Package{
		Package:     pkg,
		PackageInfo: pkgInfo,
		Registry:    rgst,
	}
	if err := p.ApplyVars(); err != nil {
		return nil, fmt.Errorf("apply the package variable: %w", err)
	}
	return p, nil
}

// getPkgInfoFromRegistries retrieves package information from the appropriate registry.
// It caches registry lookups for performance and validates package existence.
func getPkgInfoFromRegistries(logger *slog.Logger, registries map[string]*registry.Config, pkg *aqua.Package, m map[string]map[string]*registry.PackageInfo) (*registry.PackageInfo, error) {
	pkgInfoMap, ok := m[pkg.Registry]
	if !ok {
		registry, ok := registries[pkg.Registry]
		if !ok {
			return nil, errRegistryNotFound
		}
		pkgInfos := registry.PackageInfos.ToMap(logger)
		m[pkg.Registry] = pkgInfos
		pkgInfoMap = pkgInfos
	}

	pkgInfo, ok := pkgInfoMap[pkg.Name]
	if !ok {
		return nil, errPkgNotFound
	}
	return pkgInfo, nil
}
