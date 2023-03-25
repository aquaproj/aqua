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
		return fmt.Errorf("failed fs.Open dst:%s, src:%s", dst, src)
	}
	defer srcFile.Close()

	dstFile, err := fs.Create(dst)
	if err != nil {
		return fmt.Errorf("failed fs.Create dst:%s, src:%s", dst, src)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed io.Copy dst:%s, src:%s", dst, src)
	}

	return nil
}

func cpDir(fs afero.Fs, src, dst string) error {
	fileInfos, err := afero.ReadDir(fs, src)
	if err != nil {
		return fmt.Errorf("failed os.ReadDir :%w", err)
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(src, fileInfo.Name())
		dstPath := filepath.Join(dst, fileInfo.Name())

		if fileInfo.IsDir() {
			err = fs.MkdirAll(dstPath, dirPermission)
			if err != nil {
				return fmt.Errorf("failed fs.MkdirAll dst:%s :%w", dstPath, err)
			}
			err = cpDir(fs, srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("failed cpDir src:%s,dst:%s :%w", srcPath, dstPath, err)
			}
		} else {
			err = cpFile(fs, srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("failed cpFile src:%s,dst:%s :%w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

func writeDmgFile(m io.Writer, body io.Reader, dest string) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	if err != nil {
		return fmt.Errorf("read the body to (%s): %w", dest, err)
	}

	if _, err := io.Copy(m, body); err != nil {
		return fmt.Errorf("copy the body to (%s): %w", dest, err)
	}
	_, err = m.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("write the body to (%s): %w", dest, err)
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

	err = writeDmgFile(m, body, dest)
	if err != nil {
		return fmt.Errorf("writeDmgFile failed :%w", err)
	}

	exe := exec.New()
	tmpMountPoint := destDir + string(filepath.Separator) + "mount"
	_, hdiutilDetach, err := exe.HdiutilAttach(ctx, dest, tmpMountPoint)
	if err != nil {
		return fmt.Errorf("hdiutil attach failed: %w", err)
	}

	err = cpDir(fs, tmpMountPoint, destDir)
	if err != nil {
		return fmt.Errorf("failed cpDir src:%s,dst:%s :%w", tmpMountPoint, destDir, err)
	}

	err = fs.Remove(dest)
	if err != nil {
		return fmt.Errorf("failed fs.Remove :%w", err)
	}

	_, err = hdiutilDetach(ctx, exe, tmpMountPoint)
	if err != nil {
		return fmt.Errorf("failed hdiutilDetach :%w", err)
	}

	return nil
}
