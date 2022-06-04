package registry

type Config struct {
	PackageInfos PackageInfos `yaml:"packages" validate:"dive" json:"packages"`
}
