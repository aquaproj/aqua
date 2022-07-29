package domain

import "github.com/spf13/afero"

type Checksums interface {
	Get(key string) string
	Set(key, value string)
	ReadFile(fs afero.Fs, p string) error
	UpdateFile(fs afero.Fs, p string) error
}
