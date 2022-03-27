package unarchive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/util"
)

type rawUnarchiver struct {
	dest string
}

func (unarchiver *rawUnarchiver) Unarchive(body io.Reader) error {
	dest := unarchiver.dest
	if err := util.MkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0o755) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}
