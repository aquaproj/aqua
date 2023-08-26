package osfile

import "github.com/spf13/afero"

func MkdirAll(fs afero.Fs, p string) error {
	return fs.MkdirAll(p, dirPermission) //nolint:wrapcheck
}
