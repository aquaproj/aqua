package unarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type rawUnarchiver struct {
	dest string
	fs   afero.Fs
}

func (unarchiver *rawUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, prgOpts *ProgressBarOpts) error {
	dest := unarchiver.dest
	if err := util.MkdirAll(unarchiver.fs, filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := unarchiver.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
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
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}
