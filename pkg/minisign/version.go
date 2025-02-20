package minisign

import (
	_ "embed"
	"runtime"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
)

var (
	//go:embed aqua.yaml
	aquaBytes []byte
	//go:embed aqua-checksums.json
	checksumBytes []byte
	Version       string           //nolint:gochecknoglobals
	checksums     = checksum.New() //nolint:gochecknoglobals
)

func init() { //nolint:gochecknoinits
	Version = checksum.ReadEmbeddedTool(checksums, aquaBytes, checksumBytes)
	if runtime.GOOS == "darwin" && runtime.GOARCH == "amd64" {
		Version = "0.11"
	}
}

func Checksums() *checksum.Checksums {
	return checksums
}
