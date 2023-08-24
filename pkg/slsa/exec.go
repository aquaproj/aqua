package slsa

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
)

type CommandExecutor interface {
	ExecWithEnvsAndGetCombinedOutput(ctx context.Context, exePath string, args, envs []string) (string, int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error
}

type ExecutorImpl struct {
	executor        CommandExecutor
	verifierExePath string
}

func NewExecutor(executor CommandExecutor, param *config.Param) *ExecutorImpl {
	rt := runtime.NewR()
	return &ExecutorImpl{
		executor: executor,
		verifierExePath: ExePath(&ParamExePath{
			RootDir: param.RootDir,
			Runtime: rt,
		}),
	}
}

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	return m.Err
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:gomnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("Verification by slsa-verifier failed temporarily, retring")
	if err := util.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running slsa-verifier: %w", err)
	}
	return nil
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) (string, error) {
	mutex := cosign.GetMutex()
	mutex.Lock()
	defer mutex.Unlock()
	out, _, err := e.executor.ExecWithEnvsAndGetCombinedOutput(ctx, e.verifierExePath, args, nil)
	return out, err //nolint:wrapcheck
}

var errVerify = errors.New("verify with slsa-verifier")

func (e *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
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
		if _, err := e.exec(ctx, args); err == nil {
			return nil
		}
		if i == 4 { //nolint:gomnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
}
