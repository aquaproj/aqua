package installpackage

import (
	"log/slog"
	"slices"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

// filterPackagesByNames returns packages matching names.
// If names is empty, it returns pkgs as is.
// Names matching packages are added to foundNames.
func filterPackagesByNames(logger *slog.Logger, pkgs []*config.Package, names, foundNames map[string]struct{}) []*config.Package {
	if len(names) == 0 {
		return pkgs
	}
	filtered := make([]*config.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if !filterPackageByNames(pkg, names, foundNames) {
			logger.Debug("skip installing the package because the package doesn't match the specified names",
				"package_name", pkg.Package.Name,
				"package_version", pkg.Package.Version,
				"registry", pkg.Package.Registry,
			)
			continue
		}
		filtered = append(filtered, pkg)
	}
	return filtered
}

// filterPackageByNames returns true if the package should be installed.
// If names is empty, it always returns true.
// Names matching the package are added to foundNames.
func filterPackageByNames(pkg *config.Package, names, foundNames map[string]struct{}) bool {
	if len(names) == 0 {
		return true
	}
	matched := matchNames(pkg, names)
	if foundNames != nil {
		for _, name := range matched {
			foundNames[name] = struct{}{}
		}
	}
	return len(matched) > 0
}

// matchNames returns names matching the package.
// A name matches the package if it equals the package name in the configuration file,
// the package name in the registry, or one of the package's command names.
func matchNames(pkg *config.Package, names map[string]struct{}) []string {
	var matched []string
	add := func(name string) {
		if _, ok := names[name]; !ok || slices.Contains(matched, name) {
			return
		}
		matched = append(matched, name)
	}
	if pkg.Package != nil {
		add(pkg.Package.Name)
	}
	if pkg.PackageInfo != nil {
		add(pkg.PackageInfo.GetName())
		for _, file := range pkg.PackageInfo.GetFiles() {
			add(file.Name)
		}
	}
	return matched
}
