package exec

import (
	"context"
	"os"
	"testing"

	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/sirupsen/logrus"
)

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
			executor: exec.NewMock(0, nil),
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
