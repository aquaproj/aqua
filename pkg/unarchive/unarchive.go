package unarchive

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
)

var errUnsupportedFileFormat = errors.New("unsupported file format")

type coreUnarchiver interface {
	Unarchive(ctx context.Context, logger *slog.Logger, src *File) error
}

type File struct {
	Body     DownloadedFile
	Filename string
	Type     string
}

type Unarchiver struct {
	executor Executor
}

type DownloadedFile interface {
	Path() (string, error)
	ReadLast() (io.ReadCloser, error)
	Wrap(w io.Writer) io.Writer
}

func New(executor Executor) *Unarchiver {
	return &Unarchiver{
		executor: executor,
	}
}

func (u *Unarchiver) Unarchive(ctx context.Context, logger *slog.Logger, src *File, dest string) error {
	return u.getUnarchiver(logger, src, dest).Unarchive(ctx, logger, src) //nolint:wrapcheck
}

func IsUnarchived(archiveType, assetName string) bool {
	if archiveType == "raw" {
		return true
	}
	if archiveType != "" {
		return false
	}
	ext := filepath.Ext(assetName)
	return ext == "" || ext == ".exe" || ext == ".jar"
}

func (u *Unarchiver) getUnarchiver(logger *slog.Logger, src *File, dest string) coreUnarchiver {
	filename := filepath.Base(src.Filename)
	if IsUnarchived(src.Type, filename) {
		return &rawUnarchiver{
			dest: filepath.Join(dest, filename),
		}
	}
	if src.Type == "dmg" {
		return &dmgUnarchiver{
			dest:     dest,
			executor: u.executor,
		}
	}
	if src.Type == "pkg" {
		return &pkgUnarchiver{
			dest:     dest,
			executor: u.executor,
		}
	}
	switch ext := filepath.Ext(filename); ext {
	case ".dmg":
		return &dmgUnarchiver{
			dest:     dest,
			executor: u.executor,
		}
	case ".pkg":
		return &pkgUnarchiver{
			dest:     dest,
			executor: u.executor,
		}
	}

	return &handler{
		executor: u.executor,
		dest:     dest,
		filename: filename,
		logger:   logger,
	}
}
