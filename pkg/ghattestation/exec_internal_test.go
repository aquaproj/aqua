package ghattestation

import (
	"errors"
	"log/slog"
	"os/exec"
	"testing"
)

type fakeCommandExecutor struct {
	out string
	err error
}

func (e *fakeCommandExecutor) ExecStderrAndGetCombinedOutput(_ *exec.Cmd) (string, int, error) {
	return e.out, 0, e.err
}

// gh failing to authenticate doesn't fail the install, but the caller must be told
// that nothing was verified.
func TestExecutorImpl_Verify_authError(t *testing.T) {
	t.Parallel()
	e := &ExecutorImpl{
		executor: &fakeCommandExecutor{
			out: "To get started with GitHub CLI, please run:  gh auth login",
			err: errors.New("exit status 4"),
		},
		exePath: "gh",
		getEnv:  func(string) string { return "" },
		environ: func() []string { return nil },
	}
	err := e.Verify(t.Context(), slog.New(slog.DiscardHandler), &ParamVerify{
		ArtifactPath: "ar2_linux_amd64.tar.gz",
		Repository:   "aquaproj/ar2",
	})
	if !errors.Is(err, ErrSkipped) {
		t.Fatalf("error = %v, wanted ErrSkipped", err)
	}
}
