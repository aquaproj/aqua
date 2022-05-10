package exec

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestController_execCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exePath string
		args    []string
	}{
		{
			title:   "normal",
			exePath: "/bin/date",
			args:    []string{},
		},
	}
	ctrl := &Controller{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := ctrl.execCommand(ctx, d.exePath, d.args, logE)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
