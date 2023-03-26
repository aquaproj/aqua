package unarchive

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/util"
	"github.com/otiai10/copy"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const FormatDMG string = "dmg"

type dmgUnarchiver struct {
	dest     string
	executor Executor
}

type Executor interface {
	HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error)
	HdiutilDetach(ctx context.Context, mountPath string) (int, error)
}

func cpFile(fs afero.Fs, src, dst string) error {
	srcFile, err := fs.Open(src)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := fs.Create(dst)
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}

	return nil
}

func cpDirWrap(fs afero.Fs, src, dst string) error {
	if _, ok := fs.(*afero.OsFs); ok {
		return copy.Copy(src, dst) //nolint:wrapcheck
	}
	return cpDir(fs, src, dst)
}

func cpDir(fs afero.Fs, src, dst string) error {
	fileInfos, err := afero.ReadDir(fs, src)
	if err != nil {
		return fmt.Errorf("read a directory: %w", err)
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(src, fileInfo.Name())
		dstPath := filepath.Join(dst, fileInfo.Name())

		if fileInfo.IsDir() {
			if err := util.MkdirAll(fs, dstPath); err != nil {
				return fmt.Errorf("create a directory: %w", err)
			}
			if err := cpDir(fs, srcPath, dstPath); err != nil {
				return fmt.Errorf("copy a directory: %w", err)
			}
		} else {
			if err := cpFile(fs, srcPath, dstPath); err != nil {
				return fmt.Errorf("copy a file: %w", err)
			}
		}
	}

	return nil
}

func (unarchiver *dmgUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error { //nolint:cyclop
	if err := util.MkdirAll(fs, unarchiver.dest); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	tempFile, err := afero.TempFile(fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer tempFile.Close()
	defer func() {
		if err := fs.Remove(tempFile.Name()); err != nil {
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

	tmpMountPoint, err := afero.TempDir(fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}

	if _, err := unarchiver.executor.HdiutilAttach(ctx, tempFile.Name(), tmpMountPoint); err != nil {
		if err := fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporal directory created to attach a DMG file")
		}
		return fmt.Errorf("hdiutil attach: %w", err)
	}
	defer func() {
		if _, err := unarchiver.executor.HdiutilDetach(ctx, tmpMountPoint); err != nil {
			logE.WithError(err).Warn("detach a DMG file")
		}
		if err := fs.Remove(tmpMountPoint); err != nil {
			logE.WithError(err).Warn("remove a temporal directory created to attach a DMG file")
		}
	}()

	if err := cpDirWrap(fs, tmpMountPoint, unarchiver.dest); err != nil {
		return fmt.Errorf("copy a directory: %w", err)
	}

	return nil
}
