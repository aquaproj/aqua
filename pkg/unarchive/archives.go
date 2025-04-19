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

func (h *handler) HandleFile(_ context.Context, f archives.FileInfo) error {
	dstPath := filepath.Join(h.dest, h.normalizePath(f.NameInArchive))
	parentDir := filepath.Dir(dstPath)
	if err := osfile.MkdirAll(h.fs, parentDir); err != nil {
		logerr.WithError(h.logE, err).Warn("create a directory")
		return nil
	}

	if f.IsDir() {
		if err := h.fs.MkdirAll(dstPath, f.Mode()|0o700); err != nil { //nolint:mnd
			logerr.WithError(h.logE, err).Warn("create a directory")
			return nil
		}
		return nil
	}

	if f.LinkTarget != "" {
		if f.Mode()&os.ModeSymlink != 0 {
			if err := os.Symlink(f.LinkTarget, dstPath); err != nil {
				logerr.WithError(h.logE, err).WithFields(logrus.Fields{
					"link_target": f.LinkTarget,
					"link_dest":   dstPath,
				}).Warn("create a symlink")
				return nil
			}
		}
		return nil
	}

	reader, err := f.Open()
	if err != nil {
		logerr.WithError(h.logE, err).Warn("open a file")
		return nil
	}
	defer reader.Close()

	dstFile, err := h.fs.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
	if err != nil {
		logerr.WithError(h.logE, err).Warn("create a file")
		return nil
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		logerr.WithError(h.logE, err).Warn("copy a file")
		return nil
	}
	return nil
}

func (h *handler) Unarchive(ctx context.Context, _ *logrus.Entry, src *File) error {
	tempFilePath, err := src.Body.Path()
	if err != nil {
		return fmt.Errorf("get a temporary file path: %w", err)
	}
	if err := h.unarchive(ctx, src.Filename, tempFilePath); err != nil {
		return logerr.WithFields(err, logrus.Fields{ //nolint:wrapcheck
			"archived_file":     tempFilePath,
			"archived_filename": src.Filename,
		})
	}
	return nil
}

func (h *handler) normalizePath(nameInArchive string) string {
	slashCount := strings.Count(nameInArchive, "/")
	backSlashCount := strings.Count(nameInArchive, "\\")
	if backSlashCount > slashCount && filepath.Separator != '\\' {
		return strings.ReplaceAll(nameInArchive, "\\", string(filepath.Separator))
	}
	return nameInArchive
}

func (h *handler) unarchive(ctx context.Context, fileName, file string) error {
	archiveFile, err := h.fs.Open(file)
	if err != nil {
		return fmt.Errorf("open a files: %w", err)
	}
	defer archiveFile.Close()

	format, input, err := archives.Identify(ctx, fileName, archiveFile)
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
