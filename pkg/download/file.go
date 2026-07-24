package download

import (
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

type DownloadedFile struct {
	path string
	body io.ReadCloser
	pb   *progressbar.ProgressBar
}

func NewDownloadedFile(body io.ReadCloser, pb *progressbar.ProgressBar) *DownloadedFile {
	return &DownloadedFile{
		body: body,
		pb:   pb,
	}
}

func (f *DownloadedFile) Close() error {
	return f.body.Close() //nolint:wrapcheck
}

func (f *DownloadedFile) Remove() error {
	if f.path == "" {
		return nil
	}
	return os.Remove(f.path) //nolint:wrapcheck //nolint:errcheck
}

func (f *DownloadedFile) Path() (string, error) {
	if f.path != "" {
		return f.path, nil
	}
	if err := f.copy(); err != nil {
		return "", err
	}
	return f.path, nil
}

func (f *DownloadedFile) Read() (io.ReadCloser, error) {
	if f.path == "" {
		if err := f.copy(); err != nil {
			return nil, err
		}
	}
	return f.read()
}

func (f *DownloadedFile) ReadLast() (io.ReadCloser, error) {
	if f.path == "" {
		return f.body, nil
	}
	return f.read()
}

func (f *DownloadedFile) Wrap(w io.Writer) io.Writer {
	if f.pb != nil && f.path == "" {
		return io.MultiWriter(w, f.pb)
	}
	return w
}

func (f *DownloadedFile) copy() error {
	tmp, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("create a temporary file: %w", err)
	}
	f.path = tmp.Name()
	var w io.Writer = tmp
	if f.pb != nil {
		w = io.MultiWriter(tmp, f.pb)
	}
	if _, err := io.Copy(w, f.body); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}

func (f *DownloadedFile) read() (io.ReadCloser, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, fmt.Errorf("open a file: %w", err)
	}
	return file, nil
}
