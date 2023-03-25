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

func cpDir(fs afero.Fs, src, dst string) error {
	fileInfos, err := afero.ReadDir(fs, src)
	if err != nil {
		return fmt.Errorf("read a directory: %w", err)
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(src, fileInfo.Name())
		dstPath := filepath.Join(dst, fileInfo.Name())

		if fileInfo.IsDir() {
			if err := fs.MkdirAll(dstPath, dirPermission); err != nil {
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

func writeDmgFile(m io.Writer, body io.Reader, dest string) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return fmt.Errorf("read the body to (%s): %w", dest, err)
	}

	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("copy the body to (%s): %w", dest, err)
	}

	if _, err := m.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("write the body to (%s): %w", dest, err)
	}

	return nil
}

func (unarchiver *dmgUnarchiver) Unarchive(ctx context.Context, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error {
	dest := unarchiver.dest
	destDir := filepath.Dir(dest)

	if err := fs.MkdirAll(destDir, dirPermission); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	f, err := fs.OpenFile(dest, os.O_RDWR|os.O_CREATE, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("open the file: %w", err)
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

	if err := writeDmgFile(m, body, dest); err != nil {
		return fmt.Errorf("write a dmg file: %w", err)
	}

	exe := exec.New()
	tmpMountPoint := destDir + string(filepath.Separator) + "mount"
	_, hdiutilDetach, err := exe.HdiutilAttach(ctx, dest, tmpMountPoint)
	if err != nil {
		return fmt.Errorf("hdiutil attach: %w", err)
	}

	if err := cpDir(fs, tmpMountPoint, destDir); err != nil {
		return fmt.Errorf("copy a directory: %w", err)
	}

	if err := fs.Remove(dest); err != nil {
		return fmt.Errorf("remove a file: %w", err)
	}

	if _, err := hdiutilDetach(ctx, exe, tmpMountPoint); err != nil {
		return fmt.Errorf("hdiutil detach :%w", err)
	}

	return nil
}
