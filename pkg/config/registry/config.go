package registry

import (
	"github.com/sirupsen/logrus"
)

// PackageInfos represents a slice of package information.
// This is the main container for all packages in a registry.
type PackageInfos []*PackageInfo

// FormatOverride allows specifying different archive formats for specific operating systems.
// This is useful when a package provides different archive formats for different platforms.
type FormatOverride struct {
	// GOOS specifies the target operating system for this format override.
	GOOS string `json:"goos" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows" yaml:",omitempty"`
	// Format specifies the archive format to use for this operating system.
	Format string `json:"format" jsonschema:"example=tar.gz,example=raw,example=zip" yaml:",omitempty"`
}

// File represents a file to be installed from a package.
// It defines the source file within the package and how it should be installed.
type File struct {
	// Name is the name of the installed file.
	Name string `json:"name,omitempty" yaml:",omitempty"`
	// Src is the source path of the file within the package archive.
	Src string `json:"src,omitempty" yaml:",omitempty"`
	// Dir is the directory where the file should be installed.
	Dir string `json:"dir,omitempty" yaml:",omitempty"`
	// Link is the relative path from Src to the link target.
	Link string `json:"link,omitempty" yaml:",omitempty"`
	// Hard indicates whether to create a hard link instead of a symbolic link.
	Hard bool `json:"hard,omitempty" yaml:",omitempty"`
}

// ToMap converts the PackageInfos slice to a map indexed by package name.
// It includes aliases and logs conflicts at debug level.
func (p *PackageInfos) ToMap(logE *logrus.Entry) map[string]*PackageInfo {
	return p.toMap(logE, logrus.DebugLevel)
}

// toMap is the internal implementation of ToMap with configurable log level.
// It handles duplicate package names and aliases with appropriate logging.
func (p *PackageInfos) toMap(logE *logrus.Entry, logLevel logrus.Level) map[string]*PackageInfo {
	m := make(map[string]*PackageInfo, len(*p))
	logE = logE.WithField("package_name", "")
	for _, pkgInfo := range *p {
		logE := logE
		if pkgInfo == nil {
			logE.Log(logLevel, "ignore an empty package")
			continue
		}
		name := pkgInfo.GetName()
		if name == "" {
			logE.Log(logLevel, "ignore a package in the registry because the name is empty")
			continue
		}
		if _, ok := m[name]; ok {
			logE.WithField("registry_package_name", name).Log(logLevel, "ignore a package in the registry because the package name is duplicate")
			continue
		}
		m[name] = pkgInfo
		for _, alias := range pkgInfo.Aliases {
			if alias.Name == "" {
				logE.WithFields(logrus.Fields{
					"registry_package_name": name,
				}).Log(logLevel, "ignore an empty package alias in the registry")
				continue
			}
			if _, ok := m[alias.Name]; ok {
				logE.WithFields(logrus.Fields{
					"registry_package_name":  name,
					"registry_package_alias": alias,
				}).Log(logLevel, "ignore a package alias in the registry because the alias is duplicate")
				continue
			}
			m[alias.Name] = pkgInfo
		}
	}
	return m
}
