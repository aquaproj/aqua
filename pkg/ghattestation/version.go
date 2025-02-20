package ghattestation

import (
	_ "embed"
	"encoding/json"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"gopkg.in/yaml.v3"
)

var Version string

//go:embed aqua.yaml
var aquaBytes []byte

//go:embed aqua-checksums.json
var checksumBytes []byte
var checksums = checksum.New()

func init() { //nolint:gochecknoinits
	if err := json.Unmarshal(checksumBytes, checksums); err != nil {
		panic(err)
	}
	cfg := &aqua.Config{}
	if err := yaml.Unmarshal(aquaBytes, cfg); err != nil {
		panic(err)
	}
	Version = cfg.Packages[0].Version
}

func Checksums() *checksum.Checksums {
	return checksums
}
