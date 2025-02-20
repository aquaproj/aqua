package slsa

import (
	_ "embed"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

var Version string //nolint:gochecknoglobals
//go:embed aqua.yaml
var aquaBytes []byte //nolint:gochecknoglobals
//go:embed aqua-checksums.json
var checksumBytes []byte       //nolint:gochecknoglobals
var checksums = checksum.New() //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	Version = checksum.ReadEmbeddedTool(checksums, aquaBytes, checksumBytes)
}

func Checksums() *checksum.Checksums {
	return checksums
}
