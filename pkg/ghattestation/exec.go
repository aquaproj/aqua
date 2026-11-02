package ghattestation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"golang.org/x/oauth2"
)

type CommandExecutor interface {
	ExecStderrAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error)
}

type Executor interface {
	Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error
}

type ExecutorImpl struct {
	executor CommandExecutor
	exePath  string
	getEnv   func(string) string
	environ  func() []string
	// newTokenSource creates a token source getting a GitHub access token from the keyring or ghtkn.
	newTokenSource func(logger *slog.Logger) (oauth2.TokenSource, error)
	tokenSource    oauth2.TokenSource
	tokenOnce      sync.Once
	tokenErr       error
}

func NewExecutor(executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR(context.Background())
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of GitHub CLI: %w", err)
	}
	return &ExecutorImpl{
		executor:       executor,
		exePath:        exePath,
		getEnv:         os.Getenv,
		environ:        os.Environ,
		newTokenSource: github.NewTokenSource,
	}, nil
}

func wait(ctx context.Context, logger *slog.Logger, retryCount int) error {
	waitTime := time.Duration(rand.IntN(1000)) * time.Millisecond //nolint:gosec,mnd
	logger.Info("gh attestation verify failed temporarily, retrying",
		"retry_count", retryCount,
		"wait_time", waitTime)
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

func (e *ExecutorImpl) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
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
		args = append(args, "--signer-workflow", unescapeSignerWorkflow(param.SignerWorkflow))
	}
	if param.PredicateType != "" {
		args = append(args, "--predicate-type", param.PredicateType)
	}
	args = append(args, param.certificateArgs()...)
	for i := range 5 {
		err := e.exec(ctx, logger, args)
		if err == nil {
			return nil
		}
		ae := &AuthError{}
		if errors.As(err, &ae) {
			slogerr.WithError(logger, err).Warn("skip verifying GitHub Artifact Attestations because of the authentication error")
			return nil
		}
		slogerr.WithError(logger, err).Warn("execute gh attestation verify",
			"exe", e.exePath,
			"args", strings.Join(args, " "))
		if i == 4 { //nolint:mnd
			break
		}
		if err := wait(ctx, logger, i+1); err != nil {
			return err
		}
	}
	return errVerify
}

// getToken returns a GitHub access token from the keyring or ghtkn.
// It returns an empty string if neither the keyring nor ghtkn is enabled.
// The token source is created only once and the token is reused while it's valid.
func (e *ExecutorImpl) getToken(logger *slog.Logger) (string, error) {
	if e.newTokenSource == nil {
		return "", nil
	}
	e.tokenOnce.Do(func() {
		ts, err := e.newTokenSource(logger)
		if err != nil {
			e.tokenErr = err
			return
		}
		if ts != nil {
			e.tokenSource = oauth2.ReuseTokenSource(nil, ts)
		}
	})
	if e.tokenErr != nil {
		return "", e.tokenErr
	}
	if e.tokenSource == nil {
		return "", nil
	}
	token, err := e.tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("get a GitHub access token: %w", err)
	}
	return token.AccessToken, nil
}

// tokenEnv returns an environment variable to pass a GitHub access token to GitHub CLI.
// It returns an empty string if GITHUB_TOKEN or GH_TOKEN is set, because GitHub CLI reads them.
func (e *ExecutorImpl) tokenEnv(logger *slog.Logger) string {
	if e.getEnv("GITHUB_TOKEN") != "" || e.getEnv("GH_TOKEN") != "" {
		return ""
	}
	if token := e.getEnv("AQUA_GITHUB_TOKEN"); token != "" {
		return "GITHUB_TOKEN=" + token
	}
	// Pass a GitHub access token from the keyring or ghtkn.
	// Otherwise, GitHub CLI calls GitHub API without an access token
	// (GH_HOST makes GitHub CLI skip the authentication check),
	// which easily hits the rate limit of unauthenticated requests.
	token, err := e.getToken(logger)
	if err != nil {
		slogerr.WithError(logger, err).Warn("get a GitHub access token for GitHub CLI")
		return ""
	}
	if token == "" {
		return ""
	}
	return "GH_TOKEN=" + token
}

func (e *ExecutorImpl) exec(ctx context.Context, logger *slog.Logger, args []string) error {
	cmd := osexec.Command(ctx, e.exePath, args...)
	cmd.Args[0] = "gh"

	// https://github.com/aquaproj/aqua/issues/4035
	// Set GH_HOST to github.com for GitHub Enterprise
	// GitHub CLI uses the GH_HOST environment variable to determine the host for GitHub API.
	// But packages are always hosted on github.com.
	if cmd.Env == nil {
		cmd.Env = e.environ()
	}
	cmd.Env = append(cmd.Env, "GH_HOST=github.com")
	if env := e.tokenEnv(logger); env != "" {
		cmd.Env = append(cmd.Env, env)
	}

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
