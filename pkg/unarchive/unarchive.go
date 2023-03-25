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

type ProgressBarOpts struct {
	ContentLength int64
	Description   string
}

type coreUnarchiver interface {
	Unarchive(ctx context.Context, logE *logrus.Entry, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error
}

type File struct {
	Body     io.Reader
	Filename string
	Type     string
}

type UnarchiverImpl struct {
	executor Executor
}

type Unarchiver interface {
	Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error
}

type MockUnarchiver struct {
	Err error
}

func (unarchiver *MockUnarchiver) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error {
	return unarchiver.Err
}

func New(executor Executor) *UnarchiverImpl {
	return &UnarchiverImpl{
		executor: executor,
	}
}

func (unarchiver *UnarchiverImpl) Unarchive(ctx context.Context, logE *logrus.Entry, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error {
	arc, err := unarchiver.getUnarchiver(src, dest)
	if err != nil {
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	return arc.Unarchive(ctx, logE, fs, src.Body, prgOpts) //nolint:wrapcheck
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

func (unarchiver *UnarchiverImpl) getUnarchiver(src *File, dest string) (coreUnarchiver, error) {
	filename := filepath.Base(src.Filename)
	if IsUnarchived(src.Type, filename) {
		return &rawUnarchiver{
			dest: filepath.Join(dest, filename),
		}, nil
	}
	if src.Type == "dmg" {
		return &dmgUnarchiver{
			dest:     dest,
			executor: unarchiver.executor,
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
		}, nil
	case archiver.Decompressor:
		return &Decompressor{
			decompressor: t,
			dest:         filepath.Join(dest, strings.TrimSuffix(filename, filepath.Ext(filename))),
		}, nil
	}
	return nil, errUnsupportedFileFormat
}
