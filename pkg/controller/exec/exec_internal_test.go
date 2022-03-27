package exec

import (
	"context"
	"os"
	"testing"

	"github.com/aquaproj/aqua/pkg/log"
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
			exePath: "echo",
			args:    []string{"hello"},
		},
	}
	ctrl := &Controller{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		logger: log.NewLogger(""),
	}
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := ctrl.execCommand(ctx, d.exePath, d.args)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
