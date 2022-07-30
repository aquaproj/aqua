package updateaqua

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
)

const maxFileSize = 1073741824 // 1GB

var (
	errAquaNotFoundInTarball = errors.New("aqua isn't found in tarball")
	errTooBigTarBall         = errors.New("tarball is too big")
)

func unarchive(dest io.Writer, src io.Reader) error {
	zr, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("create a gzip reader: %w", err)
	}
	defer zr.Close()
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return errAquaNotFoundInTarball
		}
		if err != nil {
			return fmt.Errorf("read a tarball: %w", err)
		}
		if hdr.Name != "aqua" {
			continue
		}
		writeCount, err := io.CopyN(dest, tr, maxFileSize)
		if err == nil {
			return nil
		}
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("copy aqua: %w", err)
		}
		if writeCount >= maxFileSize {
			return errTooBigTarBall
		}
		return nil
	}
}
