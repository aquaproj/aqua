package exec

import (
	"context"
	"os"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/exec"
	"github.com/sirupsen/logrus"
)

func TestController_execCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exePath  string
		args     []string
		executor Executor
	}{
		{
			title:    "normal",
			exePath:  "/bin/date",
			args:     []string{},
			executor: &exec.Mock{},
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
			err := ctrl.execCommandWithRetry(ctx, logE, d.exePath, d.args...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
