//go:build windows
// +build windows

package exec

import "errors"

var errXSysNotSuppported = errors.New("Windows doesn't support AQUA_EXPERIMENTAL_X_SYS_EXEC")

func (exe *Executor) ExecXSysWithEnvs(exePath string, args, envs []string) error {
	return errXSysNotSuppported
}
