package checksum

import (
	"encoding/json"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"gopkg.in/yaml.v3"
)

// ReadEmbeddedTool reads embedded tool configuration and checksums from byte arrays.
// It unmarshals the checksum data into the provided Checksums struct and extracts
// the version from the first package in the aqua configuration.
// This function panics if unmarshaling fails, as it's expected to work with valid embedded data.
func ReadEmbeddedTool(checksums *Checksums, aquaBytes, checksumBytes []byte) string {
	if err := json.Unmarshal(checksumBytes, checksums); err != nil {
		panic(err)
	}
	cfg := &aqua.Config{}
	if err := yaml.Unmarshal(aquaBytes, cfg); err != nil {
		panic(err)
	}
	return cfg.Packages[0].Version
}
