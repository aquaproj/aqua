//go:build !windows
// +build !windows

package exec

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (exe *executor) ExecXSys(exePath string, args []string) error {
	return unix.Exec(exePath, append([]string{filepath.Base(exePath)}, args...), os.Environ()) //nolint:wrapcheck
}
