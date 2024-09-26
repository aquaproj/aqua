//go:build windows
// +build windows

package exec

import "errors"

var errXSysNotSuppported = errors.New("Windows doesn't support xsys")

func (e *Executor) ExecXSys(exePath string, args ...string) error {
	return errXSysNotSuppported
}
