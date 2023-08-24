package unarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Decompressor struct {
	decompressor archiver.Decompressor
	fs           afero.Fs
	dest         string
}

func (d *Decompressor) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	dest := d.dest
	if err := util.MkdirAll(d.fs, filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := d.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()

	body, err := src.Body.ReadLast()
	if err != nil {
		return fmt.Errorf("read a file: %w", err)
	}

	return d.decompressor.Decompress(body, src.Body.Wrap(f)) //nolint:wrapcheck
}
