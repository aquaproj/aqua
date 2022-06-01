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
	ExecXSys(exePath string, args []string) error
	GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error)
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}

func New() Executor {
	return &executor{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (exe *executor) ExecXSys(exePath string, args []string) error {
	return unix.Exec(exePath, append([]string{filepath.Base(exePath)}, args...), os.Environ()) //nolint:wrapcheck
}

func (exe *executor) command(cmd *exec.Cmd) *exec.Cmd {
	cmd.Stdin = exe.stdin
	cmd.Stdout = exe.stdout
	cmd.Stderr = exe.stderr
	return cmd
}

func (exe *executor) exec(ctx context.Context, cmd *exec.Cmd) (int, error) {
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

func (exe *executor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.exec(ctx, exe.command(exec.Command(exePath, args...)))
}

func (exe *executor) GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error) {
	cmd := exe.command(exec.Command("go", "build", "-o", exePath, src))
	cmd.Dir = exeDir
	return exe.exec(ctx, cmd)
}

func (exe *executor) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	cmd := exe.command(exec.Command("go", "install", path))
	cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	return exe.exec(ctx, cmd)
}
