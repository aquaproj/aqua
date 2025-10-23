package osexec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type Cmd = exec.Cmd

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

const waitDelay = 1000 * time.Hour

func (e *Executor) Exec(cmd *exec.Cmd) (int, error) {
	err := cmd.Run()
	return cmd.ProcessState.ExitCode(), err
}

func Command(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setCancel(cmd)
	return cmd
}

func setCancel(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = waitDelay
}

// ExecAndOutputWhenFailure executes a command, and outputs the command output to standard error only when the command failed.
func (e *Executor) ExecAndOutputWhenFailure(cmd *exec.Cmd) (int, error) {
	buf := &bytes.Buffer{}
	stderr := cmd.Stderr
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(stderr, buf.String())
		return cmd.ProcessState.ExitCode(), err
	}
	return 0, nil
}

func (e *Executor) ExecStderrAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error) {
	cmd.Stdout = cmd.Stderr
	return e.execAndGetCombinedOutput(cmd)
}

func (e *Executor) ExecStderr(cmd *exec.Cmd) (int, error) {
	cmd.Stdout = cmd.Stderr
	return e.Exec(cmd)
}

func (e *Executor) execAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error) {
	out := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(cmd.Stdout, out)
	cmd.Stderr = io.MultiWriter(cmd.Stderr, out)
	code, err := e.Exec(cmd)
	return out.String(), code, err
}
