package osexec_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
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
	executor := osexec.New()
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			exitCode, err := executor.Exec(osexec.Command(ctx, d.exePath, d.args...))
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
