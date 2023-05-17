package unarchive

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatPKG string = "pkg"

type pkgUnarchiver struct {
	dest     string
	executor Executor
	fs       afero.Fs
}

func (unarchiver *pkgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, prgOpts *ProgressBarOpts) error {
	if err := util.MkdirAll(unarchiver.fs, filepath.Dir(unarchiver.dest)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	tempFilePath := src.SourceFilePath
	if tempFilePath == "" {
		tmp, err := afero.TempFile(unarchiver.fs, "", "")
		if err != nil {
			return fmt.Errorf("create a temporal file: %w", err)
		}
		defer tmp.Close()
		tempFilePath = tmp.Name()
		defer func() {
			if err := unarchiver.fs.Remove(tempFilePath); err != nil {
				logE.WithError(err).Warn("remove a temporal file created to unarchive a dmg file")
			}
		}()
		var tempFile io.Writer = tmp
		if prgOpts != nil {
			bar := progressbar.DefaultBytes(
				prgOpts.ContentLength,
				prgOpts.Description,
			)
			tempFile = io.MultiWriter(tmp, bar)
		}
		if _, err := io.Copy(tempFile, src.Body); err != nil {
			return fmt.Errorf("write a dmg file: %w", err)
		}
	}

	if _, err := unarchiver.executor.UnarchivePkg(ctx, tempFilePath, unarchiver.dest); err != nil {
		return fmt.Errorf("unarchive a pkg format file: %w", err)
	}

	return nil
}
