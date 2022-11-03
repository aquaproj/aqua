package generate

import (
	"strings"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/sirupsen/logrus"
)

func excludeDuplicatedPkgs(logE *logrus.Entry, cfg *aqua.Config, pkgs []*aqua.Package) []*aqua.Package {
	ret := make([]*aqua.Package, 0, len(pkgs))
	m := make(map[string]*aqua.Package, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		pkg := pkg
		m[pkg.Registry+","+pkg.Name+"@"+pkg.Version] = pkg
		m[pkg.Registry+","+pkg.Name] = pkg
	}
	for _, pkg := range pkgs {
		pkg := pkg
		var keyV string
		var key string
		registry := registryStandard
		if pkg.Registry != "" {
			registry = pkg.Registry
		}
		if pkg.Version == "" {
			keyV = registry + "," + pkg.Name
			if pkgName, _, found := strings.Cut(pkg.Name, "@"); found {
				key = registry + "," + pkgName
			} else {
				key = keyV
			}
		} else {
			keyV = registry + "," + pkg.Name + "@" + pkg.Version
			key = registry + "," + pkg.Name
		}
		if _, ok := m[keyV]; ok {
			logE.WithFields(logrus.Fields{
				"package_name":     pkg.Name,
				"package_version":  pkg.Version,
				"package_registry": registry,
			}).Warn("skip adding a duplicate package")
			continue
		}
		m[keyV] = pkg
		ret = append(ret, pkg)
		if _, ok := m[key]; ok {
			logE.WithFields(logrus.Fields{
				"package_name":     pkg.Name,
				"package_registry": registry,
			}).Warn("same package already exists")
			continue
		}
		m[key] = pkg
	}
	return ret
}
