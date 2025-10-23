//go:build !windows
// +build !windows

package osexec

import (
	"os"

	"golang.org/x/sys/unix"
)

func (e *Executor) ExecXSys(exePath, name string, args ...string) error {
	return unix.Exec(exePath, append([]string{name}, args...), os.Environ()) //nolint:wrapcheck
}
