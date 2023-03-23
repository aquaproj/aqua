package unarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

type rawUnarchiver struct {
	dest string
}

func (unarchiver *rawUnarchiver) Unarchive(ctx context.Context, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error {
	dest := unarchiver.dest
	if err := fs.MkdirAll(filepath.Dir(dest), dirPermission); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
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

	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}
