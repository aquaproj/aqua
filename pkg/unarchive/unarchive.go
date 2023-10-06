package unarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
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

type UnarchiverImpl struct {
	executor Executor
	fs       afero.Fs
}

type DownloadedFile interface {
	Path() (string, error)
	ReadLast() (io.ReadCloser, error)
	Wrap(w io.Writer) io.Writer
}

type Unarchiver interface {
	Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string) error
}

type MockUnarchiver struct {
	Err error
}

func (u *MockUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string) error {
	return u.Err
}

func New(executor Executor, fs afero.Fs) *UnarchiverImpl {
	return &UnarchiverImpl{
		executor: executor,
		fs:       fs,
	}
}

func (u *UnarchiverImpl) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string) error {
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

func (u *UnarchiverImpl) getUnarchiver(src *File, dest string) (coreUnarchiver, error) {
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

	f := filename
	if src.Type != "" {
		f = "." + src.Type
	}
	arc, err := archiver.ByExtension(f)
	if err != nil {
		return nil, fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	switch t := arc.(type) {
	case archiver.Unarchiver:
		return &unarchiverWithUnarchiver{
			unarchiver: t,
			dest:       dest,
			fs:         u.fs,
		}, nil
	case archiver.Decompressor:
		return &Decompressor{
			decompressor: t,
			dest:         filepath.Join(dest, strings.TrimSuffix(filename, filepath.Ext(filename))),
			fs:           u.fs,
		}, nil
	}
	return nil, errUnsupportedFileFormat
}
