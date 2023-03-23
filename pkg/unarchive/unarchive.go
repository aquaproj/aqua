package unarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/afero"
)

var errUnsupportedFileFormat = errors.New("unsupported file format")

type ProgressBarOpts struct {
	ContentLength int64
	Description   string
}

type coreUnarchiver interface {
	Unarchive(ctx context.Context, fs afero.Fs, body io.Reader, prgOpts *ProgressBarOpts) error
}

type File struct {
	Body     io.Reader
	Filename string
	Type     string
}

type UnarchiverImpl struct{}

type Unarchiver interface {
	Unarchive(ctx context.Context, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error
}

type MockUnarchiver struct {
	Err error
}

func (unarchiver *MockUnarchiver) Unarchive(ctx context.Context, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error {
	return unarchiver.Err
}

func New() *UnarchiverImpl {
	return &UnarchiverImpl{}
}

func (unarchiver *UnarchiverImpl) Unarchive(ctx context.Context, src *File, dest string, fs afero.Fs, prgOpts *ProgressBarOpts) error {
	arc, err := getUnarchiver(src, dest)
	if err != nil {
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	return arc.Unarchive(ctx, fs, src.Body, prgOpts) //nolint:wrapcheck
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

func getUnarchiver(src *File, dest string) (coreUnarchiver, error) {
	filename := filepath.Base(src.Filename)
	if IsUnarchived(src.Type, filename) {
		return &rawUnarchiver{
			dest: filepath.Join(dest, filename),
		}, nil
	}
	if src.Type == "dmg" {
		return &dmgUnarchiver{
			dest: filepath.Join(dest, filename),
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
