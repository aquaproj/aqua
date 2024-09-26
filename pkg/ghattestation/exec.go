package ghattestation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type CommandExecutor interface {
	Exec(cmd *osexec.Cmd) (int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error
}

type ExecutorImpl struct {
	executor CommandExecutor
	exePath  string
}

func NewExecutor(executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR()
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of GitHub CLI: %w", err)
	}
	return &ExecutorImpl{
		executor: executor,
		exePath:  exePath,
	}, nil
}

func wait(ctx context.Context, logE *logrus.Entry, retryCount int) error {
	randGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))       //nolint:gosec
	waitTime := time.Duration(randGenerator.Intn(1000)) * time.Millisecond //nolint:mnd
	logE.WithFields(logrus.Fields{
		"retry_count": retryCount,
		"wait_time":   waitTime,
	}).Info("gh attestation verify failed temporarily, retring")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running gh attestation verify: %w", err)
	}
	return nil
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) error {
	_, err := e.executor.Exec(osexec.Command(ctx, e.exePath, args...))
	return err //nolint:wrapcheck
}

var errVerify = errors.New("verify GitHub Artifact Attestations")

func (e *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error {
	/*
		$ gh attestation verify hello \
		  -R suzuki-shunsuke/test-github-artifact-attestation \
		  --signer-workflow suzuki-shunsuke/test-github-artifact-attestation/.github/workflows/release.yaml
	*/

	args := []string{
		"attestation",
		"verify",
		param.ArtifactPath,
		"-R",
		param.Repository,
	}
	if param.SignerWorkflow != "" {
		args = append(args, "--signer-workflow", param.SignerWorkflow)
	}
	for i := range 5 {
		err := e.exec(ctx, args)
		if err == nil {
			return nil
		}
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"exe":  e.exePath,
			"args": strings.Join(args, " "),
		}).Warn("execute gh attestation verify")
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
}
