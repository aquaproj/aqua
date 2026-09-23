package slsa

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os/exec"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
)

type CommandExecutor interface {
	ExecStderrAndGetCombinedOutput(cmd *osexec.Cmd) (string, int, error)
}

type Executor interface {
	Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify, provenancePath string) error
}

type ExecutorImpl struct {
	executor        CommandExecutor
	verifierExePath string
}

func NewExecutor(executor CommandExecutor, param *config.Param) *ExecutorImpl {
	rt := runtime.NewR(context.Background())
	return &ExecutorImpl{
		executor: executor,
		verifierExePath: ExePath(&ParamExePath{
			RootDir: param.RootDir,
			Runtime: rt,
		}),
	}
}

func wait(ctx context.Context, logger *slog.Logger, retryCount int) error {
	waitTime := time.Duration(rand.IntN(1000)) * time.Millisecond //nolint:gosec,mnd
	logger.Info("Verification by slsa-verifier failed temporarily, retrying",
		"retry_count", retryCount,
		"wait_time", waitTime)
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running slsa-verifier: %w", err)
	}
	return nil
}

// ErrVerify says slsa-verifier ran and didn't accept what it was given.
//
// It is what tells a caller apart from the failures that happen before anything is
// verified -- a signature that isn't published, a file that can't be downloaded --
// which say the package isn't signed the way it was thought to be rather than that
// the signature doesn't hold. What slsa-verifier printed is wrapped with it.
var ErrVerify = errors.New("verify with slsa-verifier")

// ran reports whether slsa-verifier ran and exited, as opposed to never starting.
//
// An exit status is a verdict on what it was given. Anything else -- the executable
// isn't there, the context ended -- happened before it looked at anything,
// and saying the signature didn't hold would be answering a question nobody got to
// ask.
func ran(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

func (e *ExecutorImpl) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify, provenancePath string) error {
	if param.SourceTag == "" {
		return errors.New("source tag is empty")
	}
	args := []string{
		"verify-artifact",
		param.ArtifactPath,
		"--provenance-path",
		provenancePath,
		"--source-uri",
		param.SourceURI,
	}
	if param.SourceTag != "-" {
		args = append(args, "--source-tag", param.SourceTag)
	}
	for i := range 5 {
		_, err := e.exec(ctx, args)
		if err == nil {
			return nil
		}
		if !ran(err) {
			return fmt.Errorf("run slsa-verifier: %w", err)
		}
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logger, i+1); err != nil {
			return err
		}
	}
	return ErrVerify
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) (string, error) {
	mutex := cosign.GetMutex()
	mutex.Lock()
	defer mutex.Unlock()
	cmd := osexec.Command(ctx, e.verifierExePath, args...)
	cmd.Args[0] = "slsa-verifier"
	out, _, err := e.executor.ExecStderrAndGetCombinedOutput(cmd)
	return out, err //nolint:wrapcheck
}
