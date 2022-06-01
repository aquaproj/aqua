package exec

import (
	"context"
)

type mockExecutor struct {
	exitCode int
	err      error
}

func NewMock(exitCode int, err error) Executor {
	return &mockExecutor{
		exitCode: exitCode,
		err:      err,
	}
}

func (exe *mockExecutor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.exitCode, exe.err
}

func (exe *mockExecutor) ExecXSys(exePath string, args []string) error {
	return exe.err
}

func (exe *mockExecutor) GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error) {
	return exe.exitCode, exe.err
}

func (exe *mockExecutor) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	return exe.exitCode, exe.err
}
