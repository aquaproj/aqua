package ghattestation

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	Exec(cmd *exec.Cmd) (int, error)
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

	// ghCredential is GitHub CLI's own credential (GH_TOKEN, GITHUB_TOKEN, or `gh auth login`).
	ghCredential      *credential
	aquaCredential    *credential
	keyringCredential *credential
	ghtknCredential   *credential

	// wait waits before retrying. It's replaced in tests.
	wait func(ctx context.Context, logger *slog.Logger, retryCount int) error

	authOnce sync.Once
	// authenticated is true if GitHub CLI has its own credential.
	authenticated bool
}

func NewExecutor(executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR(context.Background())
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of GitHub CLI: %w", err)
	}
	return newExecutor(executor, exePath, os.Getenv, os.Environ, github.NewKeyringTokenSource, github.NewGhtknTokenSource), nil
}

type newTokenSource func(logger *slog.Logger) (oauth2.TokenSource, error)

func newExecutor(executor CommandExecutor, exePath string, getEnv func(string) string, environ func() []string, newKeyringTokenSource, newGhtknTokenSource newTokenSource) *ExecutorImpl {
	return &ExecutorImpl{
		executor:     executor,
		exePath:      exePath,
		getEnv:       getEnv,
		environ:      environ,
		wait:         wait,
		ghCredential: &credential{name: "gh"},
		aquaCredential: &credential{
			name: "AQUA_GITHUB_TOKEN",
			newTokenSource: func(_ *slog.Logger) (oauth2.TokenSource, error) {
				token := getEnv("AQUA_GITHUB_TOKEN")
				if token == "" {
					return nil, nil
				}
				return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}), nil
			},
		},
		keyringCredential: &credential{name: "keyring", newTokenSource: newKeyringTokenSource},
		ghtknCredential:   &credential{name: "ghtkn", newTokenSource: newGhtknTokenSource},
	}
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

var (
	errVerify       = errors.New("verify GitHub Artifact Attestations")
	errNoCredential = errors.New("no GitHub access token is available for gh attestation verify")
)

// maxRetries is the maximum number of retries for temporary errors per Verify call.
// It's shared among credentials so that the number of gh executions is bounded.
const maxRetries = 4

func (e *ExecutorImpl) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	/*
		$ gh attestation verify hello \
		  -R suzuki-shunsuke/test-github-artifact-attestation \
		  --signer-workflow suzuki-shunsuke/test-github-artifact-attestation/.github/workflows/release.yaml
	*/

	args := param.args()

	state := &verifyState{triedTokens: map[string]struct{}{}}
	for _, cred := range e.credentials(ctx) {
		token, ok := cred.token(logger)
		if !ok || state.tried(token) {
			continue
		}
		f, err := e.runWithRetry(ctx, logger, args, cred, token, state)
		if err != nil {
			return err
		}
		if f == failureNone {
			return nil
		}
		if stop := e.handleFailure(logger, cred, f, param.Private); stop {
			return errVerify
		}
		// Try the next credential.
	}
	if state.lastErr == nil {
		return errNoCredential
	}
	if state.lastFailure == failureAuth {
		// https://github.com/aquaproj/aqua/issues/3157
		// Ignore error if authentifaction fails
		slogerr.WithError(logger, &AuthError{err: state.lastErr}).Warn("skip verifying GitHub Artifact Attestations because of the authentication error")
		return nil
	}
	return errVerify
}

// args returns the arguments of gh attestation verify.
func (p *ParamVerify) args() []string {
	args := []string{
		"attestation",
		"verify",
		p.ArtifactPath,
		"-R",
		p.Repository,
	}
	if p.SignerWorkflow != "" {
		args = append(args, "--signer-workflow", unescapeSignerWorkflow(p.SignerWorkflow))
	}
	if p.PredicateType != "" {
		args = append(args, "--predicate-type", p.PredicateType)
	}
	return append(args, p.certificateArgs()...)
}

// verifyState is the state of a Verify call.
type verifyState struct {
	// retries is the number of retries for temporary errors.
	retries     int
	triedTokens map[string]struct{}
	lastErr     error
	lastFailure failure
}

// tried returns true if token has already been tried in the Verify call.
// It records token otherwise.
// An empty token (GitHub CLI's own credential) is never regarded as tried.
func (s *verifyState) tried(token string) bool {
	if token == "" {
		return false
	}
	if _, ok := s.triedTokens[token]; ok {
		return true
	}
	s.triedTokens[token] = struct{}{}
	return false
}

// runWithRetry runs gh attestation verify with token, retrying temporary errors.
// It returns failureNone if the verification succeeds.
// It returns an error if the retries are exhausted or the wait is canceled.
func (e *ExecutorImpl) runWithRetry(ctx context.Context, logger *slog.Logger, args []string, cred *credential, token string, state *verifyState) (failure, error) {
	for {
		out, err := e.run(ctx, args, token)
		if err == nil {
			return failureNone, nil
		}
		f := classify(out)
		slogerr.WithError(logger, err).Warn("execute gh attestation verify",
			"exe", e.exePath,
			"args", strings.Join(args, " "),
			"credential", cred.name,
			"failure", f.String())
		state.lastErr = err
		state.lastFailure = f
		if f != failureTemporary {
			return f, nil
		}
		if state.retries >= maxRetries {
			return f, errVerify
		}
		state.retries++
		if err := e.wait(ctx, logger, state.retries); err != nil {
			return f, err
		}
	}
}

// handleFailure handles a failure with a credential.
// It returns true if the verification should fail without trying other credentials.
func (e *ExecutorImpl) handleFailure(logger *slog.Logger, cred *credential, f failure, private bool) bool {
	switch f { //nolint:exhaustive
	case failureVerification, failureNoAttestation:
		// Changing the access token doesn't change the result.
		return true
	case failureNotFound:
		// Attestations of a public repository can be read without any access token,
		// so changing the access token doesn't change the result.
		// For a private package, 404 may be caused by the lack of the permission.
		return !private
	case failureInvalidToken, failureRateLimit:
		// The credential would fail for other packages too.
		if cred.disable() {
			logger.Warn("stop using the GitHub access token for gh attestation verify",
				"credential", cred.name,
				"failure", f.String())
		}
	}
	return false
}

// credentials returns credentials in the order of precedence.
// If GitHub CLI has its own credential, it's used first as is.
// Otherwise, GitHub CLI is run without any access token at last.
// Then GitHub CLI calls GitHub API without an access token
// (GH_HOST makes GitHub CLI skip the authentication check),
// which works for public repositories but easily hits the rate limit.
func (e *ExecutorImpl) credentials(ctx context.Context) []*credential {
	if e.isAuthenticated(ctx) {
		return []*credential{e.ghCredential, e.aquaCredential, e.keyringCredential, e.ghtknCredential}
	}
	return []*credential{e.aquaCredential, e.keyringCredential, e.ghtknCredential, e.ghCredential}
}

// isAuthenticated returns true if GitHub CLI has its own credential.
// It runs `gh auth token`, which doesn't call any API, only once.
// The output is discarded so that the access token isn't leaked.
func (e *ExecutorImpl) isAuthenticated(ctx context.Context) bool {
	e.authOnce.Do(func() {
		cmd := e.command(ctx, []string{"auth", "token", "--hostname", "github.com"}, "")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		_, err := e.executor.Exec(cmd)
		e.authenticated = err == nil
	})
	return e.authenticated
}

func (e *ExecutorImpl) command(ctx context.Context, args []string, token string) *exec.Cmd {
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
	if token != "" {
		// GH_TOKEN takes precedence over GITHUB_TOKEN and the credential stored by `gh auth login`.
		// If the same key is set multiple times, the last one is used.
		cmd.Env = append(cmd.Env, "GH_TOKEN="+token)
	}
	return cmd
}

// run runs GitHub CLI and returns the combined output and the error.
// token is passed to GitHub CLI as GH_TOKEN. It's ignored if it's empty.
func (e *ExecutorImpl) run(ctx context.Context, args []string, token string) (string, error) {
	out, _, err := e.executor.ExecStderrAndGetCombinedOutput(e.command(ctx, args, token))
	return out, err //nolint:wrapcheck
}
