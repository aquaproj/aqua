package registry

type Config struct {
	PackageInfos PackageInfos `yaml:"packages" json:"packages"`
}
