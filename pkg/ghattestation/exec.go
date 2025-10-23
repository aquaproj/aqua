package ghattestation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
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
	ExecStderrAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error)
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify) error
	VerifyRelease(ctx context.Context, logE *logrus.Entry, param *ParamVerifyRelease) error
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
	}).Info("gh attestation verify failed temporarily, retrying")
	if err := timer.Wait(ctx, waitTime); err != nil {
		return fmt.Errorf("wait running gh attestation verify: %w", err)
	}
	return nil
}

type AuthError struct {
	err error
}

func (e *AuthError) Error() string {
	return e.err.Error()
}

func (e *AuthError) Unwrap() error {
	return e.err
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
	if param.PredicateType != "" {
		args = append(args, "--predicate-type", param.PredicateType)
	}
	for i := range 5 {
		err := e.exec(ctx, args)
		if err == nil {
			return nil
		}
		ae := &AuthError{}
		if errors.As(err, &ae) {
			logerr.WithError(logE, err).Warn("skip verifying GitHub Artifact Attestations because of the authentication error")
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

type ParamVerifyRelease struct {
	ArtifactPath string
	Repository   string
	Version      string
}

func (e *ExecutorImpl) VerifyRelease(ctx context.Context, logE *logrus.Entry, param *ParamVerifyRelease) error {
	/*
		$ gh release verify-asset hello \
		  -R suzuki-shunsuke/test-github-artifact-attestation
	*/

	args := []string{
		"release",
		"verify-asset",
		"-R",
		param.Repository,
		param.Version,
		param.ArtifactPath,
	}
	for i := range 5 {
		err := e.exec(ctx, args)
		if err == nil {
			return nil
		}
		ae := &AuthError{}
		if errors.As(err, &ae) {
			logerr.WithError(logE, err).Warn("skip verifying GitHub Release Attestations because of the authentication error")
			return nil
		}
		logerr.WithError(logE, err).WithFields(logrus.Fields{
			"exe":  e.exePath,
			"args": strings.Join(args, " "),
		}).Warn("execute gh release verify-asset")
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logE, i+1); err != nil {
			return err
		}
	}
	return errVerify
}

func (e *ExecutorImpl) exec(ctx context.Context, args []string) error {
	cmd := osexec.Command(ctx, e.exePath, args...)
	cmd.Args[0] = "gh"

	// https://github.com/aquaproj/aqua/issues/4035
	// Set GH_HOST to github.com for GitHub Enterprise
	// GitHub CLI uses the GH_HOST environment variable to determine the host for GitHub API.
	// But packages are always hosted on github.com.
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "GH_HOST=github.com")

	out, _, err := e.executor.ExecStderrAndGetCombinedOutput(cmd)
	if err == nil {
		return nil
	}

	// https://github.com/aquaproj/aqua/issues/3157
	// Ignore error if authentifaction fails
	if strings.Contains(out, "gh auth login") || strings.Contains(out, "set the GH_TOKEN environment variable") {
		return &AuthError{err: err}
	}
	return err //nolint:wrapcheck
}
