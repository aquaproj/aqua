package registry

import "github.com/sirupsen/logrus"

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
