package osexec

type Mock struct {
	ExitCode int
	Err      error
	Output   string
}

func (e *Mock) Exec(cmd *Cmd) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecStderr(cmd *Cmd) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecXSys(_, _ string, _ ...string) error {
	return e.Err
}

func (e *Mock) ExecAndOutputWhenFailure(cmd *Cmd) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecStderrAndGetCombinedOutput(cmd *Cmd) (string, int, error) {
	return e.Output, e.ExitCode, e.Err
}
