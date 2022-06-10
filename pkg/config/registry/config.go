package registry

import (
	"github.com/sirupsen/logrus"
)

type PackageInfos []*PackageInfo

type FormatOverride struct {
	GOOS   string `json:"goos" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=js,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	Format string `yaml:"format" json:"format" jsonschema:"example=tar.gz,example=raw"`
}

type File struct {
	Name string `validate:"required" json:"name,omitempty"`
	Src  string `json:"src,omitempty"`
	Dir  string `json:"dir,omitempty"`
}

func (pkgInfos *PackageInfos) ToMap(logE *logrus.Entry) (map[string]*PackageInfo, error) {
	m := make(map[string]*PackageInfo, len(*pkgInfos))
	logE = logE.WithField("package_name", "")
	for _, pkgInfo := range *pkgInfos {
		logE := logE
		pkgInfo := pkgInfo
		name := pkgInfo.GetName()
		if name == "" {
			logE.Debug("ignore a package in the registry because the name is empty")
			continue
		}
		if _, ok := m[name]; ok {
			logE.WithField("registry_package_name", name).Debug("ignore a package in the registry because the package name is duplicate")
			continue
		}
		m[name] = pkgInfo
		for _, alias := range pkgInfo.Aliases {
			if alias.Name == "" {
				logE.WithFields(logrus.Fields{
					"registry_package_name": name,
				}).Debug("ignore an empty package alias in the registry")
				continue
			}
			if _, ok := m[alias.Name]; ok {
				logE.WithFields(logrus.Fields{
					"registry_package_name":  name,
					"registry_package_alias": alias,
				}).Debug("ignore a package alias in the registry because the alias is duplicate")
				continue
			}
			m[alias.Name] = pkgInfo
		}
	}
	return m, nil
}
