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

func (decompressor *Decompressor) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	dest := decompressor.dest
	if err := util.MkdirAll(decompressor.fs, filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := decompressor.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()

	body, err := src.Body.ReadLast()
	if err != nil {
		return fmt.Errorf("read a file: %w", err)
	}

	return decompressor.decompressor.Decompress(body, f) //nolint:wrapcheck
}
