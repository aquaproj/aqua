package exec

import (
	"bytes"
	"context"
	"fmt"
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

// execAndOutputWhenFailure executes a command, and outputs the command output to standard error only when the command failed.
func (exe *Executor) execAndOutputWhenFailure(ctx context.Context, cmd *exec.Cmd) (int, error) {
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	runner := timeout.NewRunner(0)
	if err := runner.Run(ctx, cmd); err != nil {
		fmt.Fprintln(exe.stderr, buf.String())
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

func (exe *Executor) ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error) {
	cmd := exec.Command(exePath, args...)
	cmd.Env = append(os.Environ(), envs...)
	return exe.exec(ctx, exe.command(cmd))
}

func (exe *Executor) ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error) {
	cmd := exe.command(exec.Command(exePath, args...))
	cmd.Env = append(os.Environ(), envs...)
	out := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(exe.stdout, out)
	cmd.Stderr = io.MultiWriter(exe.stderr, out)
	code, err := exe.exec(ctx, cmd)
	return out.String(), code, err
}

func (exe *Executor) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	cmd := exe.command(exec.Command("go", "install", path))
	cmd.Env = append(os.Environ(), "GOBIN="+gobin)
	return exe.exec(ctx, cmd)
}
