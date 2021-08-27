package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
)

func getUnarchiver(filename, typ string) (interface{}, error) {
	if typ != "" {
		filename = "." + typ
	}
	return archiver.ByExtension(filename) //nolint:wrapcheck
}

func unarchive(body io.Reader, filename, typ, dest string) error { //nolint:cyclop
	if isUnarchived(typ, filename) {
		logrus.Debug("archive type is raw")
		if err := os.MkdirAll(dest, 0o775); err != nil { //nolint:gomnd
			return fmt.Errorf("create a directory (%s): %w", dest, err)
		}
		f, err := os.OpenFile(filepath.Join(dest, filename), os.O_RDWR|os.O_CREATE, 0o755) //nolint:gomnd
		if err != nil {
			return fmt.Errorf("open the file (%s): %w", dest, err)
		}
		defer f.Close()
		if _, err := io.Copy(f, body); err != nil {
			return fmt.Errorf("copy the body to %s: %w", dest, err)
		}
		return nil
	}
	arc, err := getUnarchiver(filename, typ)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"archive_type":           typ,
			"filename":               filename,
			"filepath.Ext(filename)": filepath.Ext(filename),
		}).Error("get the unarchiver or decompressor")
		return fmt.Errorf("get the unarchiver or decompressor by the file extension: %w", err)
	}

	switch t := arc.(type) {
	case archiver.Unarchiver:
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
		return t.Unarchive(f.Name(), dest) //nolint:wrapcheck
	case archiver.Decompressor:
		f, err := os.Open(dest)
		if err != nil {
			return fmt.Errorf("open the file (%s): %w", dest, err)
		}
		defer f.Close()
		return t.Decompress(body, f) //nolint:wrapcheck
	}
	f, err := os.Open(dest)
	if err != nil {
		return fmt.Errorf("open the file (%s): %w", dest, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("copy the file to %s: %w", dest, err)
	}
	return nil
}
