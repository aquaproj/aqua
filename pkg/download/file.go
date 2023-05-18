package download

import (
	"fmt"
	"io"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
)

type DownloadedFile struct {
	fs   afero.Fs
	path string
	body io.ReadCloser
	pb   *progressbar.ProgressBar
}

func NewDownloadedFile(fs afero.Fs, body io.ReadCloser, pb *progressbar.ProgressBar) *DownloadedFile {
	return &DownloadedFile{
		fs:   fs,
		body: body,
		pb:   pb,
	}
}

func (file *DownloadedFile) Close() error {
	return file.body.Close() //nolint:wrapcheck
}

func (file *DownloadedFile) Remove() error {
	if file.path == "" {
		return nil
	}
	return file.fs.Remove(file.path) //nolint:wrapcheck //nolint:errcheck
}

func (file *DownloadedFile) GetPath() (string, error) {
	if file.path != "" {
		return file.path, nil
	}
	if err := file.copy(); err != nil {
		return "", err
	}
	return file.path, nil
}

func (file *DownloadedFile) Read() (io.ReadCloser, error) {
	if file.path == "" {
		if err := file.copy(); err != nil {
			return nil, err
		}
	}
	return file.read()
}

func (file *DownloadedFile) ReadLast() (io.ReadCloser, error) {
	if file.path == "" {
		return file.body, nil
	}
	return file.read()
}

func (file *DownloadedFile) read() (io.ReadCloser, error) {
	f, err := file.fs.Open(file.path)
	if err != nil {
		return nil, fmt.Errorf("open a file: %w", err)
	}
	return f, nil
}

func (file *DownloadedFile) Wrap(w io.Writer) io.Writer {
	if file.pb != nil && file.path == "" {
		return io.MultiWriter(w, file.pb)
	}
	return w
}

func (file *DownloadedFile) copy() error {
	tmp, err := afero.TempFile(file.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	file.path = tmp.Name()
	var w io.Writer = tmp
	if file.pb != nil {
		w = io.MultiWriter(tmp, file.pb)
	}
	if _, err := io.Copy(w, file.body); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
