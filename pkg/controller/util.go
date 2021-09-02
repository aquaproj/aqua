package controller

import "os"

func mkdirAll(p string) error {
	return os.MkdirAll(p, 0o775) //nolint:gomnd,wrapcheck
}
