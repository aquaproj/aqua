package unarchive

import (
	"context"
	"fmt"
	"io"

	"github.com/mholt/archiver/v3"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type unarchiverWithUnarchiver struct {
	unarchiver archiver.Unarchiver
	dest       string
	fs         afero.Fs
}

func (unarchiver *unarchiverWithUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, prgOpts *ProgressBarOpts) error {
	dest := unarchiver.dest
	f, err := afero.TempFile(unarchiver.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer func() {
		f.Close()
		unarchiver.fs.Remove(f.Name()) //nolint:errcheck
	}()

	var m io.Writer = f
	if prgOpts != nil {
		bar := progressbar.DefaultBytes(
			prgOpts.ContentLength,
			prgOpts.Description,
		)
		m = io.MultiWriter(f, bar)
	}

	body := src.Body
	if src.SourceFilePath != "" {
		f, err := unarchiver.fs.Open(src.SourceFilePath)
		if err != nil {
			return fmt.Errorf("open a file: %w", err)
		}
		defer f.Close()
		body = f
	}

	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("copy the file to the temporal file: %w", err)
	}
	return unarchiver.unarchiver.Unarchive(f.Name(), dest) //nolint:wrapcheck
}
