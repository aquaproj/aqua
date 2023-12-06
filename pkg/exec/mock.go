package exec

import (
	"context"
)

type Mock struct {
	ExitCode int
	Err      error
	Output   string
}

func (e *Mock) Exec(context.Context, string, ...string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecWithEnvs(context.Context, string, []string, []string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecWithEnvsAndGetCombinedOutput(context.Context, string, []string, []string) (string, int, error) {
	return e.Output, e.ExitCode, e.Err
}

func (e *Mock) ExecXSys(string, ...string) error {
	return e.Err
}

func (e *Mock) HdiutilDetach(context.Context, string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) HdiutilAttach(context.Context, string, string) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) UnarchivePkg(context.Context, string, string) (int, error) {
	return e.ExitCode, e.Err
}
