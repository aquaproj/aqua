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
	tokenMutex     sync.Mutex
	tokenSource    oauth2.TokenSource
	// tokenSourceCreated is true if newTokenSource has already been called.
	tokenSourceCreated bool
	// tokenDisabled is true if the token from the keyring or ghtkn is unavailable or invalid.
	// Then GitHub CLI is run without the token for the rest of the process.
	tokenDisabled bool
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
// If it fails to get a token, it outputs a warning and disables the token for the rest of the process,
// so that it doesn't try to get a token and output the same warning for every package and retry.
func (e *ExecutorImpl) getToken(logger *slog.Logger) string {
	e.tokenMutex.Lock()
	defer e.tokenMutex.Unlock()
	if e.tokenDisabled || e.newTokenSource == nil {
		return ""
	}
	if !e.tokenSourceCreated {
		e.tokenSourceCreated = true
		ts, err := e.newTokenSource(logger)
		if err != nil {
			e.tokenDisabled = true
			slogerr.WithError(logger, err).Warn("create a token source for GitHub CLI")
			return ""
		}
		e.tokenSource = ts
	}
	if e.tokenSource == nil {
		return ""
	}
	// Both the keyring and ghtkn token sources cache the token, so this doesn't call any API every time.
	token, err := e.tokenSource.Token()
	if err != nil {
		e.tokenDisabled = true
		slogerr.WithError(logger, err).Warn("get a GitHub access token for GitHub CLI")
		return ""
	}
	return token.AccessToken
}

// disableToken disables the token from the keyring or ghtkn for the rest of the process.
// It returns true if the token has been enabled until now.
// The return value is used to output a warning only once even if gh is run concurrently.
func (e *ExecutorImpl) disableToken() bool {
	e.tokenMutex.Lock()
	defer e.tokenMutex.Unlock()
	if e.tokenDisabled {
		return false
	}
	e.tokenDisabled = true
	return true
}

// tokenEnv returns an environment variable to pass a GitHub access token to GitHub CLI.
// It returns an empty string if GITHUB_TOKEN or GH_TOKEN is set, because GitHub CLI reads them.
// fromStore is true if the token is got from the keyring or ghtkn.
func (e *ExecutorImpl) tokenEnv(logger *slog.Logger) (env string, fromStore bool) {
	if e.getEnv("GITHUB_TOKEN") != "" || e.getEnv("GH_TOKEN") != "" {
		return "", false
	}
	if token := e.getEnv("AQUA_GITHUB_TOKEN"); token != "" {
		return "GITHUB_TOKEN=" + token, false
	}
	// Pass a GitHub access token from the keyring or ghtkn.
	// Otherwise, GitHub CLI calls GitHub API without an access token
	// (GH_HOST makes GitHub CLI skip the authentication check),
	// which easily hits the rate limit of unauthenticated requests.
	token := e.getToken(logger)
	if token == "" {
		return "", false
	}
	return "GH_TOKEN=" + token, true
}

// isTokenError returns true if GitHub CLI fails because of the token itself.
// 401 means the token is invalid (e.g. revoked or expired).
// 403 means the token lacks the permission (e.g. SAML SSO enforcement).
// 404 isn't included because it can't be distinguished from the case where attestations don't exist.
func isTokenError(out string) bool {
	return strings.Contains(out, "HTTP 401") || strings.Contains(out, "HTTP 403") || strings.Contains(out, "Bad credentials")
}

func (e *ExecutorImpl) exec(ctx context.Context, logger *slog.Logger, args []string) error {
	env, fromStore := e.tokenEnv(logger)
	out, err := e.run(ctx, args, env)
	if err == nil {
		return nil
	}
	if fromStore && isTokenError(out) {
		// GH_TOKEN takes precedence over the credential stored by `gh auth login`,
		// so the token from the keyring or ghtkn may be invalid or lack the permission GitHub CLI had before.
		// Run GitHub CLI without the token from now on, as aqua did before it passed the token.
		// Other errors (e.g. verification failures and temporary errors) are retried with the token by Verify.
		if e.disableToken() {
			slogerr.WithError(logger, err).Warn("gh attestation verify failed with the GitHub access token from the keyring or ghtkn, so run GitHub CLI without it")
		}
		out, err = e.run(ctx, args, "")
		if err == nil {
			return nil
		}
	}
	return toAuthError(out, err)
}

// run runs GitHub CLI and returns the combined output and the error.
// tokenEnv is an environment variable to pass a GitHub access token. It's ignored if it's empty.
func (e *ExecutorImpl) run(ctx context.Context, args []string, tokenEnv string) (string, error) {
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
	if tokenEnv != "" {
		cmd.Env = append(cmd.Env, tokenEnv)
	}

	out, _, err := e.executor.ExecStderrAndGetCombinedOutput(cmd)
	return out, err //nolint:wrapcheck
}

// toAuthError converts err to AuthError if GitHub CLI fails because of the authentication.
func toAuthError(out string, err error) error {
	// https://github.com/aquaproj/aqua/issues/3157
	// Ignore error if authentifaction fails
	if strings.Contains(out, "gh auth login") || strings.Contains(out, "set the GH_TOKEN environment variable") {
		return &AuthError{err: err}
	}
	return err
}
