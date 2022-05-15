package unarchive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

type rawUnarchiver struct {
	dest string
}

const (
	dirPermission  os.FileMode = 0o775
	filePermission os.FileMode = 0o755
)

func (unarchiver *rawUnarchiver) Unarchive(fs afero.Fs, body io.Reader) error {
	dest := unarchiver.dest
	if err := fs.MkdirAll(filepath.Dir(dest), dirPermission); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission)
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}
