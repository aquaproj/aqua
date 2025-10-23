package exec

import (
	"os"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/sirupsen/logrus"
)

func TestController_execCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exePath  string
		exeName  string
		args     []string
		executor Executor
	}{
		{
			title:    "normal",
			exePath:  "/bin/date",
			exeName:  "date",
			args:     []string{},
			executor: &osexec.Mock{},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			ctrl := &Controller{
				stdin:    os.Stdin,
				stdout:   os.Stdout,
				stderr:   os.Stderr,
				executor: d.executor,
			}
			err := ctrl.execCommandWithRetry(ctx, logE, d.exePath, d.exeName, d.args...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
