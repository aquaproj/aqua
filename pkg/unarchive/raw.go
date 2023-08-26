package unarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type rawUnarchiver struct {
	dest string
	fs   afero.Fs
}

func (u *rawUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	dest := u.dest
	if err := osfile.MkdirAll(u.fs, filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := u.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()

	body, err := src.Body.ReadLast()
	if err != nil {
		return fmt.Errorf("read a file: %w", err)
	}
	if _, err := io.Copy(src.Body.Wrap(f), body); err != nil {
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}
