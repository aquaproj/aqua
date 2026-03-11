package exec

import (
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
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
	logger := slog.New(slog.DiscardHandler)
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
			err := ctrl.execCommandWithRetry(ctx, logger, d.exePath, d.exeName, d.args...)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_wrapExec(t *testing.T) {
	t.Parallel()
	data := []struct {
		title      string
		lookPath   func(exeName string) (string, error)
		exeName    string
		exePath    string
		args       []string
		expExeName string
		expExePath string
		expArgs    []string
		isErr      bool
	}{
		{
			title:      "non jar",
			exeName:    "foo",
			exePath:    "/usr/bin/foo",
			args:       []string{"--help"},
			expExeName: "foo",
			expExePath: "/usr/bin/foo",
			expArgs:    []string{"--help"},
		},
		{
			title: "jar",
			lookPath: func(string) (string, error) {
				return "/usr/bin/java", nil
			},
			exeName:    "app",
			exePath:    "/path/to/app.jar",
			args:       []string{"arg1"},
			expExeName: "java",
			expExePath: "/usr/bin/java",
			expArgs:    []string{"-jar", "/path/to/app.jar", "arg1"},
		},
		{
			title: "jar without args",
			lookPath: func(string) (string, error) {
				return "/usr/bin/java", nil
			},
			exeName:    "app",
			exePath:    "/path/to/app.jar",
			args:       []string{},
			expExeName: "java",
			expExePath: "/usr/bin/java",
			expArgs:    []string{"-jar", "/path/to/app.jar"},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			exeName, exePath, args, err := wrapExec(d.lookPath, d.exeName, d.exePath, d.args...)
			if err != nil {
				if !d.isErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if d.isErr {
				t.Fatalf("expected error, but got none")
			}
			if exeName != d.expExeName {
				t.Fatalf("exeName = %s, want %s", exeName, d.expExeName)
			}
			if exePath != d.expExePath {
				t.Fatalf("exePath = %s, want %s", exePath, d.expExePath)
			}
			if !slices.Equal(args, d.expArgs) {
				t.Fatalf("args = %v, want %v", args, d.expArgs)
			}
		})
	}
}
