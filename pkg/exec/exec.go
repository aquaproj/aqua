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

type Cmd = exec.Cmd

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

const waitDelay = 1000 * time.Hour

type ParamRun struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (e *Executor) Exec(cmd *exec.Cmd, param *ParamRun) (int, error) {
	if param != nil && param.Stdout != nil {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, param.Stdout)
	}
	if param != nil && param.Stderr != nil {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, param.Stderr)
	}
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

// execAndOutputWhenFailure executes a command, and outputs the command output to standard error only when the command failed.
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

func (e *Executor) ExecAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error) {
	out := &bytes.Buffer{}
	code, err := e.Exec(cmd, &ParamRun{
		Stdout: out,
		Stderr: out,
	})
	return out.String(), code, err
}
