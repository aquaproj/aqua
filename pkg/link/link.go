package link

import (
	"errors"
	"os"

	"github.com/spf13/afero"
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

func (f *mockFileInfo) Mode() os.FileMode {
	return os.ModeSymlink
}

type mockLinker struct {
	files map[string]*mockFileInfo
	fs    afero.Fs
}

func NewMockLinker(fs afero.Fs) Linker {
	return &mockLinker{
		files: map[string]*mockFileInfo{},
		fs:    fs,
	}
}

func (lk *mockLinker) Lstat(s string) (os.FileInfo, error) {
	if f, ok := lk.files[s]; ok {
		return f, nil
	}
	return lk.fs.Stat(s) //nolint:wrapcheck
}

func (lk *mockLinker) Symlink(dest, src string) error {
	if _, ok := lk.files[src]; ok {
		return errors.New("file already exists")
	}
	if _, err := lk.fs.Create(src); err != nil {
		return err //nolint:wrapcheck
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
