package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
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

func (exe *Executor) ExecCommand(cmd *exec.Cmd) (int, error) {
	return exe.exec(exe.command(cmd))
}

func (exe *Executor) Exec(ctx context.Context, exePath string, args ...string) (int, error) {
	return exe.exec(exe.command(exec.CommandContext(ctx, exePath, args...)))
}

func (exe *Executor) ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error) {
	cmd := exec.CommandContext(ctx, exePath, args...)
	cmd.Env = append(os.Environ(), envs...)
	return exe.exec(exe.command(cmd))
}

func (exe *Executor) ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error) {
	cmd := exe.command(exec.CommandContext(ctx, exePath, args...))
	cmd.Env = append(os.Environ(), envs...)
	out := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(exe.stdout, out)
	cmd.Stderr = io.MultiWriter(exe.stderr, out)
	code, err := exe.exec(cmd)
	return out.String(), code, err
}

func (exe *Executor) command(cmd *exec.Cmd) *exec.Cmd {
	cmd.Stdin = exe.stdin
	cmd.Stdout = exe.stdout
	cmd.Stderr = exe.stderr
	return cmd
}

const waitDelay = 1000 * time.Hour

func setCancel(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt) //nolint:wrapcheck
	}
	cmd.WaitDelay = waitDelay
}

func (exe *Executor) exec(cmd *exec.Cmd) (int, error) {
	setCancel(cmd)
	if err := cmd.Run(); err != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

// execAndOutputWhenFailure executes a command, and outputs the command output to standard error only when the command failed.
func (exe *Executor) execAndOutputWhenFailure(cmd *exec.Cmd) (int, error) {
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	setCancel(cmd)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(exe.stderr, buf.String())
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}
