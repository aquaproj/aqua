package unarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/mholt/archiver/v3"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Decompressor struct {
	decompressor archiver.Decompressor
	fs           afero.Fs
	dest         string
}

func (decompressor *Decompressor) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, prgOpts *ProgressBarOpts) error {
	dest := decompressor.dest
	if err := util.MkdirAll(decompressor.fs, filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := decompressor.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
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
		f, err := decompressor.fs.Open(src.SourceFilePath)
		if err != nil {
			return fmt.Errorf("open a file: %w", err)
		}
		defer f.Close()
		body = f
	}

	return decompressor.decompressor.Decompress(body, m) //nolint:wrapcheck
}
