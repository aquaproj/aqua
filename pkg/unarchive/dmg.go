package unarchive

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatDMG string = "dmg"

type dmgUnarchiver struct {
	dest     string
	executor Executor
	fs       afero.Fs
}

type Executor interface {
	HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error)
	HdiutilDetach(ctx context.Context, mountPath string) (int, error)
}

func (unarchiver *dmgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, body io.Reader, prgOpts *ProgressBarOpts) error { //nolint:cyclop
	if err := util.MkdirAll(unarchiver.fs, unarchiver.dest); err != nil {
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

	tmpMountPoint, err := afero.TempDir(unarchiver.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}

	if _, err := unarchiver.executor.HdiutilAttach(ctx, tempFile.Name(), tmpMountPoint); err != nil {
		if err := unarchiver.fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporal directory created to attach a DMG file")
		}
		return fmt.Errorf("hdiutil attach: %w", err)
	}
	defer func() {
		if _, err := unarchiver.executor.HdiutilDetach(ctx, tmpMountPoint); err != nil {
			logE.WithError(err).Warn("detach a DMG file")
		}
		if err := unarchiver.fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporal directory created to attach a DMG file")
		}
	}()

	if err := cpDirWrap(unarchiver.fs, tmpMountPoint, unarchiver.dest); err != nil {
		return fmt.Errorf("copy a directory: %w", err)
	}

	return nil
}
