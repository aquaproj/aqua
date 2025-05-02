package registry

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	PackageInfos PackageInfos `yaml:"packages" json:"packages"`
	m            map[string]*PackageInfo
}

func (c *Config) Packages(logE *logrus.Entry) map[string]*PackageInfo {
	if c.m != nil {
		return c.m
	}
	m := c.PackageInfos.ToMap(logE)
	c.m = m
	return m
}

func (c *Config) Package(logE *logrus.Entry, pkgName string) *PackageInfo {
	m := c.Packages(logE)
	pkg, ok := m[pkgName]
	if !ok {
		return nil
	}
	return pkg
}
