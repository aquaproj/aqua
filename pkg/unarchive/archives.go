package unarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/mholt/archives"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type handler struct {
	fs   afero.Fs
	dest string
}

const readOnlyPerm = 0o200

func allowWrite(fs afero.Fs, path string) (func() error, error) {
	originalMode, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat a parent directory: %w", err)
	}

	if originalMode.Mode().Perm()&readOnlyPerm != 0 {
		return nil, nil
	}

	if err := os.Chmod(path, originalMode.Mode()|readOnlyPerm); err != nil {
		return nil, fmt.Errorf("chmod parent directory: %w", err)
	}
	return func() error {
		return fs.Chmod(path, originalMode.Mode())
	}, nil
}

func (h *handler) HandleFile(_ context.Context, f archives.FileInfo) error {
	dstPath := filepath.Clean(filepath.Join(h.dest, f.NameInArchive))

	parentDir := filepath.Dir(dstPath)
	if err := osfile.MkdirAll(h.fs, parentDir); err != nil {
		return err
	}

	if f.IsDir() {
		if err := h.fs.MkdirAll(dstPath, f.Mode()); err != nil {
			return fmt.Errorf("create a directory: %w", err)
		}
		return nil
	}

	if f.LinkTarget != "" {
		return nil
	}

	fn, err := allowWrite(h.fs, parentDir)
	if err != nil {
		return err
	}
	if fn != nil {
		defer fn()
	}

	reader, err := f.Open()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer reader.Close()

	dstFile, err := h.fs.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}

func (h *handler) Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error {
	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	return h.unarchive(ctx, tempFilePath)
}

func (h *handler) unarchive(ctx context.Context, tarball string) error {
	archiveFile, err := h.fs.Open(tarball)
	if err != nil {
		return fmt.Errorf("open tarball %s: %w", tarball, err)
	}
	defer archiveFile.Close()

	format, input, err := archives.Identify(ctx, tarball, archiveFile)
	if err != nil {
		return fmt.Errorf("identify the format: %w", err)
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return errors.New("the file format isn't supported")
	}

	if err := osfile.MkdirAll(h.fs, h.dest); err != nil {
		return fmt.Errorf("create a destination directory: %w", err)
	}

	if err := extractor.Extract(ctx, input, h.HandleFile); err != nil {
		return fmt.Errorf("extract files: %w", err)
	}

	return nil
}
