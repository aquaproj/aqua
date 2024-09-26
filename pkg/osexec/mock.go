package osexec

type Mock struct {
	ExitCode int
	Err      error
	Output   string
}

func (e *Mock) Exec(cmd *Cmd) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecXSys(exePath string, args ...string) error {
	return e.Err
}

func (e *Mock) ExecAndOutputWhenFailure(cmd *Cmd) (int, error) {
	return e.ExitCode, e.Err
}

func (e *Mock) ExecAndGetCombinedOutput(cmd *Cmd) (string, int, error) {
	return e.Output, e.ExitCode, e.Err
}
