//go:build windows
// +build windows

package osexec

import "errors"

var errXSysNotSupported = errors.New("Windows doesn't support xsys")

func (e *Executor) ExecXSys(exePath string, args ...string) error {
	return errXSysNotSupported
}
