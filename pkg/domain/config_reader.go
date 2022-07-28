package domain

import (
	"github.com/aquaproj/aqua/pkg/config/aqua"
)

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}
