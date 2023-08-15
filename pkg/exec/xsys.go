//go:build !windows
// +build !windows

package exec

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (exe *Executor) ExecXSysWithEnvs(exePath string, args, envs []string) error {
	return unix.Exec(exePath, append([]string{filepath.Base(exePath)}, args...), append(os.Environ(), envs...)) //nolint:wrapcheck
}
