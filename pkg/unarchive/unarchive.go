package unarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var errUnsupportedFileFormat = errors.New("unsupported file format")

type coreUnarchiver interface {
	Unarchive(ctx context.Context, logE *logrus.Entry, src *File) error
}

type File struct {
	Body     DownloadedFile
	Filename string
	Type     string
}

type Unarchiver struct {
	executor Executor
	fs       afero.Fs
}

type DownloadedFile interface {
	Path() (string, error)
	ReadLast() (io.ReadCloser, error)
	Wrap(w io.Writer) io.Writer
}

func New(executor Executor, fs afero.Fs) *Unarchiver {
	return &Unarchiver{
		executor: executor,
		fs:       fs,
	}
}

func (u *Unarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string) error {
	arc, err := u.getUnarchiver(src, dest)
	if err != nil {
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	return arc.Unarchive(ctx, logE, src) //nolint:wrapcheck
}

func IsUnarchived(archiveType, assetName string) bool {
	if archiveType == "raw" {
		return true
	}
	if archiveType != "" {
		return false
	}
	ext := filepath.Ext(assetName)
	return ext == "" || ext == ".exe"
}

func (u *Unarchiver) getUnarchiver(src *File, dest string) (coreUnarchiver, error) {
	filename := filepath.Base(src.Filename)
	if IsUnarchived(src.Type, filename) {
		return &rawUnarchiver{
			dest: filepath.Join(dest, filename),
			fs:   u.fs,
		}, nil
	}
	if src.Type == "dmg" {
		return &dmgUnarchiver{
			dest:     dest,
			executor: u.executor,
			fs:       u.fs,
		}, nil
	}
	if src.Type == "pkg" {
		return &pkgUnarchiver{
			dest:     dest,
			executor: u.executor,
			fs:       u.fs,
		}, nil
	}
	switch ext := filepath.Ext(filename); ext {
	case ".dmg":
		return &dmgUnarchiver{
			dest:     dest,
			executor: u.executor,
			fs:       u.fs,
		}, nil
	case ".pkg":
		return &pkgUnarchiver{
			dest:     dest,
			executor: u.executor,
			fs:       u.fs,
		}, nil
	}

	return &handler{
		fs:   u.fs,
		dest: dest,
	}, nil
}
