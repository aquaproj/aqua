package exec

import (
	"context"
)

type Mock struct {
	ExitCode int
	Err      error
	Output   string
}

func (e *Mock) Exec(ctx context.Context, exePath string, args ...string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error) {
	return e.Output, e.ExitCode, e.Err
}

func (e *Mock) ExecXSys(exePath string, args ...string) error {
	return e.Err
}

func (e *Mock) HdiutilDetach(ctx context.Context, mountPath string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) HdiutilAttach(ctx context.Context, dmgPath, mountPoint string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) UnarchivePkg(ctx context.Context, pkgFilePath, dest string) (int, error) {
	return e.ExitCode, e.Err
}
