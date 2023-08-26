package osfile

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/spf13/afero"
)

func Copy(fs afero.Fs, src, dst string) error {
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
			if err := MkdirAll(fs, dstPath); err != nil {
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
