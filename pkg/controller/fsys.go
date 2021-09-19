package controller

import (
	"io"
	"os"
)

type Fsys interface {
	Stat(string) (os.FileInfo, error)
	Lstat(string) (os.FileInfo, error)
	Chmod(string, os.FileMode) error
	Symlink(string, string) error
	Readlink(string) (string, error)
	Remove(string) error
	Getwd(string) (string, error)
	Open(string) (io.ReadCloser, error)
	OpenFile(string, os.FileMode) (io.WriteCloser, error)
	WriteFile(string, []byte, os.FileMode) error
	CreateTemp(string, string) (io.WriteCloser, error)
}
