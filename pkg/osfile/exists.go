package osfile

import (
	"errors"
	"fmt"
	"os"
)

// Exists reports whether the file or directory p exists.
//
// A file that exists but can't be stat'd, for instance because a directory in
// the path isn't readable, is not a missing file: the error is returned rather
// than reported as false. Only os.ErrNotExist becomes (false, nil). Callers
// that genuinely don't care why the stat failed, such as walking up a directory
// tree for the nearest configuration file, discard the error deliberately.
func Exists(p string) (bool, error) {
	if _, err := os.Stat(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("check if a file exists: %w", err)
	}
	return true, nil
}
