package unarchive

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatPKG string = "pkg"

type pkgUnarchiver struct {
	dest     string
	executor Executor
	fs       afero.Fs
}

func (unarchiver *pkgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	if err := util.MkdirAll(unarchiver.fs, filepath.Dir(unarchiver.dest)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath, err := src.Body.GetPath()
	if err != nil {
		return fmt.Errorf("get a temporal file path: %w", err)
	}

	if _, err := unarchiver.executor.UnarchivePkg(ctx, tempFilePath, unarchiver.dest); err != nil {
		return fmt.Errorf("unarchive a pkg format file: %w", err)
	}

	return nil
}
