//go:build windows

package osexec

import "errors"

var errXSysNotSupported = errors.New("Windows doesn't support xsys")

func (e *Executor) ExecXSys(_, _ string, _ ...string) error {
	return errXSysNotSupported
}
