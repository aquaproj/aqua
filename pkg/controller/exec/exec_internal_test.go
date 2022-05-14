package exec

import (
	"context"
	"os"
	"testing"

	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/sirupsen/logrus"
)

type mockExecutor struct {
	exitCode int
	err      error
}

func (exe *mockExecutor) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.exitCode, exe.err
}

func (exe *mockExecutor) ExecXSys(exePath string, args []string) error {
	return exe.err
}

func TestController_execCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exePath  string
		args     []string
		executor exec.Executor
	}{
		{
			title:    "normal",
			exePath:  "/bin/date",
			args:     []string{},
			executor: &mockExecutor{},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			ctrl := &Controller{
				stdin:    os.Stdin,
				stdout:   os.Stdout,
				stderr:   os.Stderr,
				executor: d.executor,
			}
			err := ctrl.execCommandWithRetry(ctx, d.exePath, d.args, logE)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
