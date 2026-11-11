package installpackage

import (
	"log/slog"
	"slices"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

// filterTargetsByNames drops the targets whose package matches none of the names.
//
// The names are matched against the package rather than the target, which is the package
// plus what installing it needs, so this is the same filter as filterPackagesByNames and
// the one to use where what is installed is a list of targets.
func filterTargetsByNames(logger *slog.Logger, targets []*Target, names, foundNames map[string]struct{}) []*Target {
	if len(names) == 0 {
		return targets
	}
	filtered := make([]*Target, 0, len(targets))
	for _, target := range targets {
		pkg := target.Pkg
		if !filterPackageByNames(pkg, names, foundNames) {
			logger.Debug("skip installing the package because the package doesn't match the specified names",
				"package_name", pkg.Package.Name,
				"package_version", pkg.Package.Version,
				"registry", pkg.Package.Registry,
			)
			continue
		}
		filtered = append(filtered, target)
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
