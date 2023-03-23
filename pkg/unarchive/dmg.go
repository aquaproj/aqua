package unarchive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

const FormatDMG string = "dmg"

type dmgUnarchiver struct {
	dest string
}

func cpFile(fs afero.Fs, src, dst string) error {
	srcFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func copyDir(fs afero.Fs, src, dst string) error {
	fileInfos, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(src, fileInfo.Name())
		dstPath := filepath.Join(dst, fileInfo.Name())

		if fileInfo.IsDir() {
			err = fs.MkdirAll(dstPath, dirPermission)
			if err != nil {
				return err
			}
			err = copyDir(fs, srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = cpFile(fs, srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (unarchiver *dmgUnarchiver) Unarchive(ctx context.Context, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error {
	dest := unarchiver.dest
	destDir := filepath.Dir(dest)

	if err := fs.MkdirAll(destDir, dirPermission); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
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

	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("copy the body to (%s): %w", dest, err)
	}
	f.Write(buf.Bytes())

	exe := exec.New()
	tmpMountPoint := destDir + string(filepath.Separator) + "mount"
	code, hdiutilDetach, err := exe.HdiutilAttach(ctx, dest, tmpMountPoint)
	if err != nil || code != 0 {
		return fmt.Errorf("hdiutil attach failed: %w", err)
	}
	defer hdiutilDetach(ctx, exe, tmpMountPoint)

	copyDir(fs, tmpMountPoint, destDir)
	fs.Remove(dest)

	return nil
}
