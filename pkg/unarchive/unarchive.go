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

	f := filename
	if src.Type != "" {
		f = "." + src.Type
	}
	arc, err := byExtension(f)
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

func byExtension(filename string) (interface{}, error) {
	formats := map[string]interface{}{
		"tbr":  archiver.NewTarBrotli(),
		"tbz":  archiver.NewTarBz2(),
		"tbz2": archiver.NewTarBz2(),
		"tgz":  archiver.NewTarGz(),
		"tlz4": archiver.NewTarLz4(),
		"tsz":  archiver.NewTarSz(),
		"txz":  archiver.NewTarXz(),
	}
	for format, arc := range formats {
		if strings.HasSuffix(filename, "."+format) {
			return arc, nil
		}
	}
	return archiver.ByExtension(filename) //nolint:wrapcheck
}
