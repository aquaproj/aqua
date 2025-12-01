package registry

import (
	"github.com/sirupsen/logrus"
)

// Config represents a registry configuration containing package definitions.
// It provides methods to access packages by name with caching for performance.
type Config struct {
	// PackageInfos contains all package definitions in the registry.
	PackageInfos PackageInfos `json:"packages" yaml:"packages"`
	// m is an internal cache of packages indexed by name (including aliases).
	m map[string]*PackageInfo
}

// Packages returns a map of all packages indexed by name (including aliases).
// The result is cached for subsequent calls to improve performance.
func (c *Config) Packages(logE *logrus.Entry) map[string]*PackageInfo {
	if c.m != nil {
		return c.m
	}
	m := c.PackageInfos.ToMap(logE)
	c.m = m
	return m
}

// Package returns the PackageInfo for the specified package name.
// It returns nil if the package is not found in the registry.
func (c *Config) Package(logE *logrus.Entry, pkgName string) *PackageInfo {
	m := c.Packages(logE)
	pkg, ok := m[pkgName]
	if !ok {
		return nil
	}
	return pkg
}
