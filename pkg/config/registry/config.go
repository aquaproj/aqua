package registry

import (
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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

func (pkgInfos *PackageInfos) ToMap() (map[string]*PackageInfo, error) {
	m := make(map[string]*PackageInfo, len(*pkgInfos))
	for _, pkgInfo := range *pkgInfos {
		pkgInfo := pkgInfo
		name := pkgInfo.GetName()
		if _, ok := m[name]; ok {
			return nil, logerr.WithFields(errPkgNameMustBeUniqueInRegistry, logrus.Fields{ //nolint:wrapcheck
				"package_name": name,
			})
		}
		m[name] = pkgInfo
		for _, alias := range pkgInfo.Aliases {
			if _, ok := m[alias.Name]; ok {
				return nil, logerr.WithFields(errPkgNameMustBeUniqueInRegistry, logrus.Fields{ //nolint:wrapcheck
					"package_name": alias.Name,
				})
			}
			m[alias.Name] = pkgInfo
		}
	}
	return m, nil
}
