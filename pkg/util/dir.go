package util

import "os"

func MkdirAll(p string) error {
	return os.MkdirAll(p, 0o775) //nolint:gomnd,wrapcheck
}
