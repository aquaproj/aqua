package domain

import (
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/security"
)

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type SecurityConfigReader interface {
	Read(configFilePath string, cfg *security.Config) error
}
