package osfile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

func MkdirAll(fs afero.Fs, p string) error {
	return fs.MkdirAll(p, dirPermission) //nolint:wrapcheck
}

// MkdirTemp creates a directory with a random name in dir and returns its path.
// The parent directory dir must exist.
//
// It is afero.TempDir with the permissions aqua grants its own directories.
// afero.TempDir hardcodes 0700 and ignores the umask, which is too strict for a
// directory that is later renamed into place and read by other users.
func MkdirTemp(fs afero.Fs, dir string) (string, error) {
	b := make([]byte, 16) //nolint:mnd
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate a random directory name: %w", err)
	}
	// Mkdir rather than MkdirAll: a name that is already taken, by a leftover
	// directory or by another user in a shared root directory, must be an error
	// rather than a directory shared with whatever created it.
	p := filepath.Join(dir, hex.EncodeToString(b))
	if err := fs.Mkdir(p, dirPermission); err != nil {
		return "", fmt.Errorf("create a directory: %w", err)
	}
	return p, nil
}
