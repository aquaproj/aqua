package util

import "github.com/spf13/afero"

func Exist(fs afero.Fs, path string) (bool, error) {
	if _, err := fs.Stat(path); err != nil {
		return false, nil //nolint:nilerr
	}
	return true, nil
}

func ExistFile(fs afero.Fs, path string) (bool, error) {
	return Exist(fs, path)
}

func ExistDir(fs afero.Fs, path string) (bool, error) {
	return Exist(fs, path)
}
