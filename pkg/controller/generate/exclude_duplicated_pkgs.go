package generate

import (
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func excludeDuplicatedPkgs(logger *slog.Logger, cfg *aqua.Config, pkgs []*config.Package) []*config.Package {
	ret := make([]*config.Package, 0, len(pkgs))
	m := make(map[string]struct{}, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		m[pkg.Registry+","+pkg.Name+"@"+pkg.Version] = struct{}{}
		m[pkg.Registry+","+pkg.Name] = struct{}{}
	}
	for _, pkg := range pkgs {
		var keyV string
		var key string
		registry := registryStandard
		if pkg.Package.Registry != "" {
			registry = pkg.Package.Registry
		}
		if pkg.Package.Version == "" {
			keyV = registry + "," + pkg.Package.Name
			if pkgName, _, found := strings.Cut(pkg.Package.Name, "@"); found {
				key = registry + "," + pkgName
			} else {
				key = keyV
			}
		} else {
			keyV = registry + "," + pkg.Package.Name + "@" + pkg.Package.Version
			key = registry + "," + pkg.Package.Name
		}
		if _, ok := m[keyV]; ok {
			logger.Warn("skip adding a duplicate package",
				"package_name", pkg.Package.Name,
				"package_version", pkg.Package.Version,
				"package_registry", registry)
			continue
		}
		m[keyV] = struct{}{}
		ret = append(ret, pkg)
		if _, ok := m[key]; ok {
			logger.Warn("same package already exists",
				"package_name", pkg.Package.Name,
				"package_registry", registry)
			continue
		}
		m[key] = struct{}{}
	}
	return ret
}
