package exec

import (
	"context"
)

type MockExecutor struct {
	ExitCode int
	Err      error
}

func (exe *MockExecutor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.ExitCode, exe.Err
}

func (exe *MockExecutor) ExecXSys(exePath string, args []string) error {
	return exe.Err
}

func (exe *MockExecutor) GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error) {
	return exe.ExitCode, exe.Err
}

func (exe *MockExecutor) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	return exe.ExitCode, exe.Err
}
