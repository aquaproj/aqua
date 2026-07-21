package osfile

import "os"

// Exists returns true if the file or directory exists.
// An error other than os.ErrNotExist, such as a permission error, is treated as
// if the file didn't exist, because aqua reads or creates the file right after
// this check anyway and would fail there with a better message.
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
