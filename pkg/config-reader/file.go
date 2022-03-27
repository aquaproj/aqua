package reader

import (
	"io"
	"os"
)

type FileReader interface {
	Read(p string) (io.ReadCloser, error)
}

func NewFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct{}

func (reader *fileReader) Read(p string) (io.ReadCloser, error) {
	return os.Open(p) //nolint:wrapcheck
}
