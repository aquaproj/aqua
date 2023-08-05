package installpackage

import (
	"fmt"
	"io"
	"os"
)

const (
	executableFilePermission os.FileMode = 0o755
)

func (inst *InstallerImpl) Copy(dest, src string) error {
	dst, err := inst.fs.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, executableFilePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dst.Close()
	srcFile, err := inst.fs.Open(src)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer srcFile.Close()
	if _, err := io.Copy(dst, srcFile); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}

	return nil
}
