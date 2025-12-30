package registry

import (
	"log/slog"
)

// PackageInfos represents a slice of package information.
// This is the main container for all packages in a registry.
type PackageInfos []*PackageInfo

// FormatOverride allows specifying different archive formats for specific operating systems.
// This is useful when a package provides different archive formats for different platforms.
type FormatOverride struct {
	// GOOS specifies the target operating system for this format override.
	GOOS string `yaml:",omitempty" json:"goos" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	// Format specifies the archive format to use for this operating system.
	Format string `yaml:",omitempty" json:"format" jsonschema:"example=tar.gz,example=raw,example=zip"`
}

// File represents a file to be installed from a package.
// It defines the source file within the package and how it should be installed.
type File struct {
	// Name is the name of the installed file.
	Name string `yaml:",omitempty" json:"name,omitempty"`
	// Src is the source path of the file within the package archive.
	Src string `yaml:",omitempty" json:"src,omitempty"`
	// Dir is the directory where the file should be installed.
	Dir string `yaml:",omitempty" json:"dir,omitempty"`
	// Link is the relative path from Src to the link target.
	Link string `yaml:",omitempty" json:"link,omitempty"`
	// Hard indicates whether to create a hard link instead of a symbolic link.
	Hard bool `yaml:",omitempty" json:"hard,omitempty"`
}

// ToMap converts the PackageInfos slice to a map indexed by package name.
// It includes aliases and logs conflicts at debug level.
func (p *PackageInfos) ToMap(logger *slog.Logger) map[string]*PackageInfo {
	return p.toMap(logger, slog.LevelDebug)
}

// toMap is the internal implementation of ToMap with configurable log level.
// It handles duplicate package names and aliases with appropriate logging.
func (p *PackageInfos) toMap(logger *slog.Logger, logLevel slog.Level) map[string]*PackageInfo {
	m := make(map[string]*PackageInfo, len(*p))
	for _, pkgInfo := range *p {
		if pkgInfo == nil {
			logger.Log(nil, logLevel, "ignore an empty package") //nolint:staticcheck
			continue
		}
		name := pkgInfo.GetName()
		if name == "" {
			logger.Log(nil, logLevel, "ignore a package in the registry because the name is empty") //nolint:staticcheck
			continue
		}
		if _, ok := m[name]; ok {
			logger.Log(nil, logLevel, "ignore a package in the registry because the package name is duplicate", "registry_package_name", name) //nolint:staticcheck
			continue
		}
		m[name] = pkgInfo
		for _, alias := range pkgInfo.Aliases {
			if alias.Name == "" {
				logger.Log(nil, logLevel, "ignore an empty package alias in the registry", "registry_package_name", name) //nolint:staticcheck
				continue
			}
			if _, ok := m[alias.Name]; ok {
				logger.Log(nil, logLevel, "ignore a package alias in the registry because the alias is duplicate", //nolint:staticcheck
					"registry_package_name", name,
					"registry_package_alias", alias,
				)
				continue
			}
			m[alias.Name] = pkgInfo
		}
	}
	return m
}
