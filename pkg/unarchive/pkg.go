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

func (unarchiver *pkgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, body io.Reader, prgOpts *ProgressBarOpts) error {
	if err := util.MkdirAll(unarchiver.fs, filepath.Dir(unarchiver.dest)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	tempFile, err := afero.TempFile(unarchiver.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer tempFile.Close()
	defer func() {
		if err := unarchiver.fs.Remove(tempFile.Name()); err != nil {
			logE.WithError(err).Warn("remove a temporal file created to unarchive a dmg file")
		}
	}()

	var m io.Writer = tempFile
	if prgOpts != nil {
		bar := progressbar.DefaultBytes(
			prgOpts.ContentLength,
			prgOpts.Description,
		)
		m = io.MultiWriter(tempFile, bar)
	}

	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("write a dmg file: %w", err)
	}

	if _, err := unarchiver.executor.UnarchivePkg(ctx, tempFile.Name(), unarchiver.dest); err != nil {
		return fmt.Errorf("unarchive a pkg format file: %w", err)
	}

	return nil
}
