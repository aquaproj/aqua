package slsa

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/util"
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
	mutex           *sync.Mutex
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
		mutex: &sync.Mutex{},
	}
}

type MockExecutor struct {
	Err error
}

func (mock *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	return mock.Err
}

const tempErrMsg = "resource temporarily unavailable"

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

func (exe *ExecutorImpl) exec(ctx context.Context, args []string) (string, error) {
	exe.mutex.Lock()
	defer exe.mutex.Unlock()
	out, _, err := exe.executor.ExecWithEnvsAndGetCombinedOutput(ctx, exe.verifierExePath, args, nil)
	return out, err //nolint:wrapcheck
}

var errVerify = errors.New("verify with slsa-verifier")

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
		out, err := exe.exec(ctx, args)
		if err == nil {
			return nil
		}
		if !strings.Contains(out, tempErrMsg) {
			return fmt.Errorf("verify with slsa-verifier: %w", err)
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
