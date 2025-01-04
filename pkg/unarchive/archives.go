package unarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/mholt/archives"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type handler struct {
	fs       afero.Fs
	dest     string
	filename string
	logE     *logrus.Entry
}

const readOnlyPerm = 0o200

func allowWrite(fs afero.Fs, path string) (func() error, error) {
	originalMode, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat a parent directory: %w", err)
	}

	if originalMode.Mode().Perm()&readOnlyPerm != 0 {
		return nil, nil //nolint:nilnil
	}

	if err := os.Chmod(path, originalMode.Mode()|readOnlyPerm); err != nil {
		return nil, fmt.Errorf("chmod parent directory: %w", err)
	}
	return func() error {
		return fs.Chmod(path, originalMode.Mode())
	}, nil
}

func (h *handler) normalizePath(nameInArchive string) string {
	slashCount := strings.Count(nameInArchive, "/")
	backSlashCount := strings.Count(nameInArchive, "\\")
	if backSlashCount > slashCount && filepath.Separator != '\\' {
		return strings.ReplaceAll(nameInArchive, "\\", string(filepath.Separator))
	}
	return nameInArchive
}

func (h *handler) HandleFile(_ context.Context, f archives.FileInfo) error {
	dstPath := filepath.Join(h.dest, h.normalizePath(f.NameInArchive))
	parentDir := filepath.Dir(dstPath)
	if err := osfile.MkdirAll(h.fs, parentDir); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}

	if f.IsDir() {
		if err := h.fs.MkdirAll(dstPath, f.Mode()); err != nil {
			return fmt.Errorf("create a directory: %w", err)
		}
		return nil
	}

	// if f.LinkTarget != "" {
	// 	return nil
	// }

	fn, err := allowWrite(h.fs, parentDir)
	if err != nil {
		return err
	}
	if fn != nil {
		defer func() {
			if err := fn(); err != nil {
				logerr.WithError(h.logE, err).Warn("failed to restore the original permission")
			}
		}()
	}

	reader, err := f.Open()
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
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

func (h *handler) Unarchive(ctx context.Context, _ *logrus.Entry, src *File) error {
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

	if extractor, ok := format.(archives.Extractor); ok {
		if err := osfile.MkdirAll(h.fs, h.dest); err != nil {
			return fmt.Errorf("create a destination directory: %w", err)
		}

		if err := extractor.Extract(ctx, input, h.HandleFile); err != nil {
			return fmt.Errorf("extract files: %w", err)
		}
		return nil
	}
	if decomp, ok := format.(archives.Decompressor); ok {
		return h.decompress(input, decomp)
	}
	return errUnsupportedFileFormat
}

func (h *handler) decompress(input io.Reader, decomp archives.Decompressor) error {
	rc, err := decomp.OpenReader(input)
	if err != nil {
		return fmt.Errorf("open a decompressed file: %w", err)
	}
	defer rc.Close()
	if err := osfile.MkdirAll(h.fs, h.dest); err != nil {
		return fmt.Errorf("create a directory (%s): %w", h.dest, err)
	}
	dst, err := h.fs.Create(filepath.Join(h.dest, strings.TrimSuffix(h.filename, filepath.Ext(h.filename))))
	if err != nil {
		return fmt.Errorf("create a destination file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, rc); err != nil {
		return fmt.Errorf("copy decompressed data: %w", err)
	}
	return nil
}
