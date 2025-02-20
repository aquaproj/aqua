package minisign

import (
	_ "embed"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

var Version string

//go:embed aqua.yaml
var aquaBytes []byte

//go:embed aqua-checksums.json
var checksumBytes []byte
var checksums = checksum.New()

func init() { //nolint:gochecknoinits
	Version = checksum.ReadEmbeddedTool(checksums, aquaBytes, checksumBytes)
}

func Checksums() *checksum.Checksums {
	return checksums
}
