package unarchive

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
)

var errUnsupportedFileFormat = errors.New("unsupported file format")

type Unarchiver interface {
	Unarchive(body io.Reader) error
}

func Unarchive(body io.Reader, filename, typ, dest string, logE *logrus.Entry) error {
	arc, err := getUnarchiver(filename, typ, dest)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"format":                 typ,
			"filename":               filename,
			"filepath.Ext(filename)": filepath.Ext(filename),
		}).Error("get the unarchiver or decompressor")
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	return arc.Unarchive(body) //nolint:wrapcheck
}

func IsUnarchived(archiveType, assetName string) bool {
	return archiveType == "raw" || (archiveType == "" && filepath.Ext(assetName) == "")
}

func getUnarchiver(filename, typ, dest string) (Unarchiver, error) {
	filename = filepath.Base(filename)
	if IsUnarchived(typ, filename) {
		return &rawUnarchiver{
			dest: filepath.Join(dest, filename),
		}, nil
	}

	f := filename
	if typ != "" {
		f = "." + typ
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
