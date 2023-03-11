package slsa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
)

type CommandExecutor interface {
	Exec(ctx context.Context, exePath string, args []string) (int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error
}

type ExecutorImpl struct {
	executor        CommandExecutor
	verifierExePath string
}

func NewExecutor(executor CommandExecutor) *ExecutorImpl {
	return &ExecutorImpl{
		executor: executor,
	}
}

type MockExecutor struct {
	Err error
}

func (mock *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	return mock.Err
}

func (exe *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	args := []string{
		"verify-artifact",
		param.ArtifactPath,
		"--provenance-path",
		provenancePath,
		"--source-uri",
		param.SourceURI,
		"--source-tag",
		param.SourceTag,
	}
	for i := 0; i < 5; i++ {
		_, err := exe.executor.Exec(ctx, exe.verifierExePath, args)
		if err == nil {
			return nil
		}
		if e := ctx.Err(); e != nil {
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		if i == 4 { //nolint:gomnd
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		logE.WithField("retry_count", i+1).Info("slsa-verifier failed. Retrying")
		if err := util.Wait(ctx, 1*time.Second); err != nil {
			return err //nolint:wrapcheck
		}
	}
	return nil
}
