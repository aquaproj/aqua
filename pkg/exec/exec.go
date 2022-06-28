package exec

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

type Executor struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func New() *Executor {
	return &Executor{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (exe *Executor) command(cmd *exec.Cmd) *exec.Cmd {
	cmd.Stdin = exe.stdin
	cmd.Stdout = exe.stdout
	cmd.Stderr = exe.stderr
	return cmd
}

func (exe *Executor) exec(ctx context.Context, cmd *exec.Cmd) (int, error) {
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

func (exe *Executor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.exec(ctx, exe.command(exec.Command(exePath, args...)))
}

func (exe *Executor) GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error) {
	cmd := exe.command(exec.Command("go", "build", "-o", exePath, src))
	cmd.Dir = exeDir
	return exe.exec(ctx, cmd)
}

func (exe *Executor) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	cmd := exe.command(exec.Command("go", "install", path))
	cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	return exe.exec(ctx, cmd)
}
