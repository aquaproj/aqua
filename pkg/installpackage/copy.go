package installpackage

import (
	"fmt"
	"io"
	"os"
)

const (
	executableFilePermission os.FileMode = 0o755
)

func (is *Installer) Copy(dest, src string) (gErr error) {
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, executableFilePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	// Close exactly once, in the deferred call. A failure to flush the
	// executable to disk must not be reported as a successful install, but it
	// must not overwrite an earlier error either.
	defer func() {
		if err := dst.Close(); err != nil && gErr == nil {
			gErr = fmt.Errorf("close the destination file: %w", err)
		}
	}()
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer srcFile.Close()
	if _, err := io.Copy(dst, srcFile); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
