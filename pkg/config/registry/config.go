package registry

import (
	"github.com/sirupsen/logrus"
)

type PackageInfos []*PackageInfo

type FormatOverride struct {
	GOOS   string `yaml:",omitempty" json:"goos" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	Format string `yaml:",omitempty" json:"format" jsonschema:"example=tar.gz,example=raw,example=zip"`
}

type File struct {
	Name string `json:"name,omitempty" yaml:",omitempty"`
	Src  string `json:"src,omitempty" yaml:",omitempty"`
	Dir  string `json:"dir,omitempty" yaml:",omitempty"`
	// Link is the relative path from Src to the link
	Link string `json:"link,omitempty" yaml:",omitempty"`
	Hard bool   `json:"hard,omitempty" yaml:",omitempty"`
}

func (p *PackageInfos) ToMap(logE *logrus.Entry) map[string]*PackageInfo {
	return p.toMap(logE, logrus.DebugLevel)
}

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
