package checksum

import (
	"encoding/json"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"gopkg.in/yaml.v3"
)

func ReadEmbeddedTool(checksums *Checksums, aquaBytes, checksumBytes []byte) string {
	err := json.Unmarshal(checksumBytes, checksums)
	if err != nil {
		panic(err)
	}

	cfg := &aqua.Config{}
	err := yaml.Unmarshal(aquaBytes, cfg)
	if err != nil {
		panic(err)
	}

	return cfg.Packages[0].Version
}
