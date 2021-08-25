package controller

import (
	"fmt"
	"io"
	"os"

	"github.com/mholt/archiver/v3"
)

func getUnarchiver(filename, typ string) (interface{}, error) {
	if typ != "" {
		filename = "." + typ
	}
	return archiver.ByExtension(filename) //nolint:wrapcheck
}

func unarchive(body io.Reader, filename, typ, dest string) error {
	arc, err := getUnarchiver(filename, typ)
	if err != nil {
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
