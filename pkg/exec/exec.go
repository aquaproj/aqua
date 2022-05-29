package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/suzuki-shunsuke/go-timeout/timeout"
	"golang.org/x/sys/unix"
)

type executor struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type Executor interface {
	Exec(ctx context.Context, exePath string, args []string) (int, error)
	ExecWithDir(ctx context.Context, exePath string, args []string, dir string) (int, error)
	ExecXSys(exePath string, args []string) error
}

func New() Executor {
	return &executor{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

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

func (exe *mockExecutor) ExecWithDir(ctx context.Context, exePath string, args []string, dir string) (int, error) {
	return exe.exitCode, exe.err
}

func (exe *mockExecutor) ExecXSys(exePath string, args []string) error {
	return exe.err
}

func (exe *executor) ExecXSys(exePath string, args []string) error {
	return unix.Exec(exePath, append([]string{filepath.Base(exePath)}, args...), os.Environ()) //nolint:wrapcheck
}

type Result struct {
	ExitCode int
}

func (exe *executor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = exe.stdin
	cmd.Stdout = exe.stdout
	cmd.Stderr = exe.stderr
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

func (exe *executor) ExecWithDir(ctx context.Context, exePath string, args []string, dir string) (int, error) {
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = exe.stdin
	cmd.Stdout = exe.stdout
	cmd.Stderr = exe.stderr
	cmd.Dir = dir
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}
