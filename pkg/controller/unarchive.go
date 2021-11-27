package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
)

type Unarchiver interface {
	Unarchive(body io.Reader) error
}

func isUnarchived(archiveType, assetName string) bool {
	return archiveType == "raw" || (archiveType == "" && filepath.Ext(assetName) == "")
}

func getUnarchiver(filename, typ, dest string) (Unarchiver, error) {
	filename = filepath.Base(filename)
	if isUnarchived(typ, filename) {
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

type rawUnarchiver struct {
	dest string
}

func (unarchiver *rawUnarchiver) Unarchive(body io.Reader) error {
	dest := unarchiver.dest
	if err := mkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0o755) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the body to %s: %w", dest, err)
	}
	return nil
}

type unarchiverWithUnarchiver struct {
	unarchiver archiver.Unarchiver
	dest       string
}

func (unarchiver *unarchiverWithUnarchiver) Unarchive(body io.Reader) error {
	dest := unarchiver.dest
	f, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the file to the temporal file: %w", err)
	}
	return unarchiver.unarchiver.Unarchive(f.Name(), dest) //nolint:wrapcheck
}

type Decompressor struct {
	decompressor archiver.Decompressor
	dest         string
}

func (decomressor *Decompressor) Unarchive(body io.Reader) error {
	dest := decomressor.dest
	if err := mkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("create a directory (%s): %w", dest, err)
	}
	f, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0o755) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	return decomressor.decompressor.Decompress(body, f) //nolint:wrapcheck
}

func unarchive(body io.Reader, filename, typ, dest string) error {
	arc, err := getUnarchiver(filename, typ, dest)
	if err != nil {
		log.New().WithFields(logrus.Fields{
			"format":                 typ,
			"filename":               filename,
			"filepath.Ext(filename)": filepath.Ext(filename),
		}).Error("get the unarchiver or decompressor")
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	return arc.Unarchive(body) //nolint:wrapcheck
}
