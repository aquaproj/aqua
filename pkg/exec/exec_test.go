package exec_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/exec"
)

func TestExecutorExec(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		exePath  string
		args     []string
		isErr    bool
		exitCode int
	}{
		{
			name:    "/bin/date",
			exePath: "/bin/date",
		},
	}
	executor := exec.New()
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			exitCode, err := executor.Exec(ctx, d.exePath, d.args...)
			if d.isErr {
				if err == nil {
					t.Fatal("err should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d.exitCode != exitCode {
				t.Fatalf("wanted %v, got %v", d.exitCode, exitCode)
			}
		})
	}
}
