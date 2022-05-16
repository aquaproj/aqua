package link

import (
	"errors"
	"os"
)

type Linker interface {
	Lstat(s string) (os.FileInfo, error)
	Symlink(dest, src string) error
	Readlink(src string) (string, error)
}

type linker struct{}

func New() Linker {
	return &linker{}
}

func (lk *linker) Lstat(s string) (os.FileInfo, error) {
	return os.Lstat(s) //nolint:wrapcheck
}

func (lk *linker) Symlink(dest, src string) error {
	return os.Symlink(dest, src) //nolint:wrapcheck
}

func (lk *linker) Readlink(src string) (string, error) {
	return os.Readlink(src) //nolint:wrapcheck
}

type mockFileInfo struct {
	os.FileInfo
	Dest string
}

type mockLinker struct {
	files map[string]*mockFileInfo
}

func NewMockLinker() Linker {
	return &mockLinker{
		files: map[string]*mockFileInfo{},
	}
}

func (lk *mockLinker) Lstat(s string) (os.FileInfo, error) {
	if f, ok := lk.files[s]; ok {
		return f, nil
	}
	return nil, errors.New("file isn't found")
}

func (lk *mockLinker) Symlink(dest, src string) error {
	if _, ok := lk.files[src]; ok {
		return errors.New("file already exists")
	}
	if lk.files == nil {
		lk.files = map[string]*mockFileInfo{}
	}
	lk.files[src] = &mockFileInfo{
		Dest: dest,
	}
	return nil
}

func (lk *mockLinker) Readlink(src string) (string, error) {
	if f, ok := lk.files[src]; ok {
		return f.Dest, nil
	}
	return "", errors.New("file isn't found")
}
