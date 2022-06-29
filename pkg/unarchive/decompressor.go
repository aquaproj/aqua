package unarchive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

type Decompressor struct {
	decompressor archiver.Decompressor
	dest         string
}

func (decompressor *Decompressor) Unarchive(fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error {
	dest := decompressor.dest
	if err := fs.MkdirAll(filepath.Dir(dest), dirPermission); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission)
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()

	var m io.Writer = f
	if prgOpts != nil {
		bar := progressbar.DefaultBytes(
			prgOpts.ContentLength,
			prgOpts.Description,
		)
		m = io.MultiWriter(f, bar)
	}

	return decompressor.decompressor.Decompress(body, m) //nolint:wrapcheck
}
