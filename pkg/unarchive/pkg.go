package unarchive

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatPKG string = "pkg"

type pkgUnarchiver struct {
	dest     string
	executor Executor
	fs       afero.Fs
}

func (u *pkgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	if err := osfile.MkdirAll(u.fs, filepath.Dir(u.dest)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporal file path: %w", err)
	}

	if _, err := u.executor.UnarchivePkg(ctx, tempFilePath, u.dest); err != nil {
		return fmt.Errorf("unarchive a pkg format file: %w", err)
	}

	return nil
}
