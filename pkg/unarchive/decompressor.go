package unarchive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/util"
	"github.com/mholt/archiver/v3"
)

type Decompressor struct {
	decompressor archiver.Decompressor
	dest         string
}

func (decomressor *Decompressor) Unarchive(body io.Reader) error {
	dest := decomressor.dest
	if err := util.MkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0o755) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	return decomressor.decompressor.Decompress(body, f) //nolint:wrapcheck
}
