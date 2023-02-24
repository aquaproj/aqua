package link

import (
	"os"
)

type Linker struct{}

func New() *Linker {
	return &Linker{}
}

func (lk *Linker) Lstat(s string) (os.FileInfo, error) {
	return os.Lstat(s) //nolint:wrapcheck
}

func (lk *Linker) Symlink(dest, src string) error {
	return os.Symlink(dest, src) //nolint:wrapcheck
}

func (lk *Linker) Readlink(src string) (string, error) {
	return os.Readlink(src) //nolint:wrapcheck
}
