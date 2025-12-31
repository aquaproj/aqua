package registry

import (
	"log/slog"
)

// Config represents a registry configuration containing package definitions.
// It provides methods to access packages by name with caching for performance.
type Config struct {
	// PackageInfos contains all package definitions in the registry.
	PackageInfos PackageInfos `yaml:"packages" json:"packages"`
	// m is an internal cache of packages indexed by name (including aliases).
	m map[string]*PackageInfo
}

// Packages returns a map of all packages indexed by name (including aliases).
// The result is cached for subsequent calls to improve performance.
func (c *Config) Packages(logger *slog.Logger) map[string]*PackageInfo {
	if c.m != nil {
		return c.m
	}
	m := c.PackageInfos.ToMap(logger)
	c.m = m
	return m
}

// Package returns the PackageInfo for the specified package name.
// It returns nil if the package is not found in the registry.
func (c *Config) Package(logger *slog.Logger, pkgName string) *PackageInfo {
	m := c.Packages(logger)
	pkg, ok := m[pkgName]
	if !ok {
		return nil
	}
	return pkg
}
