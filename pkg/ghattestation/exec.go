package ghattestation

import (
	"bytes"
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
}

type Executor interface {
	Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error
}

type ExecutorImpl struct {
	executor CommandExecutor
	exePath  string
	getEnv   func(string) string
	environ  func() []string

	aquaCredential    *credential
	keyringCredential *credential
	ghtknCredential   *credential

	// stderr is where the output of GitHub CLI is written. It's replaced in tests.
	stderr io.Writer

	// wait waits before retrying. It's replaced in tests.
	wait func(ctx context.Context, logger *slog.Logger, retryCount int) error

	authOnce sync.Once
	// authenticated is true if GitHub CLI has its own credential (GH_TOKEN, GITHUB_TOKEN, or `gh auth login`).
	authenticated bool
	// ghTokenKey is the key of GitHub CLI's own access token.
	ghTokenKey string

	disabledMutex sync.Mutex
	// disabledKeys are keys of access tokens which would fail for other packages too (e.g. invalid tokens and rate limits).
	// They aren't used for the rest of the process regardless of their sources.
	disabledKeys map[string]struct{}
}

func NewExecutor(executor CommandExecutor, param *config.Param) (*ExecutorImpl, error) {
	rt := runtime.NewR(context.Background())
	pkg := Package()
	pkg.PackageInfo.OverrideByRuntime(rt)
	exePath, err := pkg.ExePath(param.RootDir, pkg.PackageInfo.GetFiles()[0], rt)
	if err != nil {
		return nil, fmt.Errorf("get an executable file path of GitHub CLI: %w", err)
	}
	newKeyringTokenSource := func(logger *slog.Logger) (oauth2.TokenSource, error) {
		return github.NewKeyringTokenSource(logger), nil
	}
	return newExecutor(executor, exePath, os.Getenv, os.Environ, newKeyringTokenSource, github.NewGhtknTokenSource), nil
}

type newTokenSource func(logger *slog.Logger) (oauth2.TokenSource, error)

func newExecutor(executor CommandExecutor, exePath string, getEnv func(string) string, environ func() []string, newKeyringTokenSource, newGhtknTokenSource newTokenSource) *ExecutorImpl {
	return &ExecutorImpl{
		executor:     executor,
		exePath:      exePath,
		getEnv:       getEnv,
		environ:      environ,
		stderr:       os.Stderr,
		wait:         wait,
		disabledKeys: map[string]struct{}{},
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

// docVerify is the document of errors of the verification.
// It describes how to solve them, such as passing a GitHub access token.
const docVerify = "https://aquaproj.github.io/docs/reference/codes/008"

var (
	errVerify       = slogerr.With(errors.New("verify GitHub Artifact Attestations"), "doc", docVerify)
	errNoCredential = slogerr.With(errors.New("no GitHub access token is available for gh attestation verify"), "doc", docVerify)
)

// maxRetries is the maximum number of retries for temporary errors per Verify call.
// It's shared among credentials so that the number of gh executions is bounded.
// If it's exhausted, the remaining credentials are tried once each without retries.
const maxRetries = 4

func (e *ExecutorImpl) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	/*
		$ gh attestation verify hello \
		  -R suzuki-shunsuke/test-github-artifact-attestation \
		  --signer-workflow suzuki-shunsuke/test-github-artifact-attestation/.github/workflows/release.yaml
	*/

	args := param.args()

	state := &verifyState{triedKeys: map[string]struct{}{}}
	err := e.verify(ctx, logger, args, param, state)
	// The output of GitHub CLI is written after all attempts,
	// so that the failures of attempts before a successful one aren't shown.
	if err == nil {
		fmt.Fprint(e.stderr, state.successOutput)
		return nil
	}
	for _, out := range state.outputs {
		fmt.Fprint(e.stderr, out)
	}
	return err
}

func (e *ExecutorImpl) verify(ctx context.Context, logger *slog.Logger, args []string, param *ParamVerify, state *verifyState) error {
	for _, getCandidate := range e.candidates(ctx) {
		cand, ok := getCandidate(logger)
		if !ok || e.isDisabled(cand.key) || state.tried(cand.key) {
			continue
		}
		f, err := e.runWithRetry(ctx, logger, args, cand, state)
		if err != nil {
			return err
		}
		if f == failureNone {
			return nil
		}
		if stop := e.handleFailure(logger, cand, f, param.Private); stop {
			return errVerify //nolint:wrapcheck
		}
		// Try the next credential.
	}
	if !state.executed {
		return errNoCredential //nolint:wrapcheck
	}
	// The verification isn't skipped even if all credentials fail because of the authentication.
	// Otherwise, packages would be installed without the verification.
	return errVerify //nolint:wrapcheck
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
	retries int
	// triedKeys are keys of access tokens tried in the Verify call.
	triedKeys map[string]struct{}
	// executed is true if gh attestation verify has been run at least once.
	executed bool
	// outputs are the outputs of failed attempts.
	outputs []string
	// successOutput is the output of the successful attempt.
	successOutput string
}

// tried returns true if the access token has already been tried in the Verify call.
// It records the key otherwise.
func (s *verifyState) tried(key string) bool {
	if _, ok := s.triedKeys[key]; ok {
		return true
	}
	s.triedKeys[key] = struct{}{}
	return false
}

// runWithRetry runs gh attestation verify with token, retrying temporary errors.
// It returns failureNone if the verification succeeds.
// If the retries are exhausted, it returns failureTemporary so that the next credential is tried.
// It returns an error if the wait is canceled.
func (e *ExecutorImpl) runWithRetry(ctx context.Context, logger *slog.Logger, args []string, cand candidate, state *verifyState) (failure, error) {
	for {
		out, err := e.run(ctx, args, cand.token)
		if err == nil {
			state.successOutput = out
			return failureNone, nil
		}
		state.outputs = append(state.outputs, out)
		f := classify(out)
		slogerr.WithError(logger, err).Warn("execute gh attestation verify",
			"exe", e.exePath,
			"args", strings.Join(args, " "),
			"credential", cand.name,
			"failure", f.String())
		state.executed = true
		if f != failureTemporary {
			return f, nil
		}
		if state.retries >= maxRetries {
			return f, nil
		}
		state.retries++
		if err := e.wait(ctx, logger, state.retries); err != nil {
			return f, err
		}
	}
}

// handleFailure handles a failure with a credential.
// It returns true if the verification should fail without trying other credentials.
func (e *ExecutorImpl) handleFailure(logger *slog.Logger, cand candidate, f failure, private bool) bool {
	switch f { //nolint:exhaustive
	case failureVerification, failureNoAttestation:
		// Changing the access token doesn't change the result.
		return true
	case failureUnknown:
		// It may be deterministic, so other credentials aren't tried.
		return true
	case failureNotFound:
		// Attestations of a public repository can be read without any access token,
		// so changing the access token doesn't change the result.
		// For a private package, 404 may be caused by the lack of the permission.
		return !private
	case failureInvalidToken, failureRateLimit:
		// The access token would fail for other packages too.
		if e.disable(cand.key) {
			logger.Warn("stop using the GitHub access token for gh attestation verify",
				"credential", cand.name,
				"failure", f.String())
		}
	}
	return false
}

// isDisabled returns true if the access token is disabled for the rest of the process.
func (e *ExecutorImpl) isDisabled(key string) bool {
	e.disabledMutex.Lock()
	defer e.disabledMutex.Unlock()
	_, ok := e.disabledKeys[key]
	return ok
}

// disable disables the access token for the rest of the process.
// It returns true if the access token has been enabled until now.
// The return value is used to output a warning only once even if gh is run concurrently.
func (e *ExecutorImpl) disable(key string) bool {
	e.disabledMutex.Lock()
	defer e.disabledMutex.Unlock()
	if _, ok := e.disabledKeys[key]; ok {
		return false
	}
	e.disabledKeys[key] = struct{}{}
	return true
}

type getCandidate func(logger *slog.Logger) (candidate, bool)

// candidates returns functions getting candidates in the order of precedence.
// They are called lazily, so later sources such as ghtkn aren't touched if an earlier one succeeds.
// If GitHub CLI has its own credential, it's used first as is.
// Otherwise, GitHub CLI is run without any access token at last.
// Then GitHub CLI calls GitHub API without an access token
// (GH_HOST makes GitHub CLI skip the authentication check),
// which works for public repositories but easily hits the rate limit.
func (e *ExecutorImpl) candidates(ctx context.Context) []getCandidate {
	authenticated := e.isAuthenticated(ctx)
	gh := func(_ *slog.Logger) (candidate, bool) {
		return candidate{name: "gh", key: e.ghTokenKey}, true
	}
	if authenticated {
		return []getCandidate{gh, e.aquaCredential.candidate, e.keyringCredential.candidate, e.ghtknCredential.candidate}
	}
	return []getCandidate{e.aquaCredential.candidate, e.keyringCredential.candidate, e.ghtknCredential.candidate, gh}
}

// ghKeyAnonymous is the key of running GitHub CLI without any access token.
const ghKeyAnonymous = "gh:anonymous"

// isAuthenticated returns true if GitHub CLI has its own credential.
// It runs `gh auth token`, which doesn't call any API, only once.
// It also records the hash of the access token to compare it with access tokens from other sources.
// The access token is read into memory instead of the terminal, and it isn't output.
func (e *ExecutorImpl) isAuthenticated(ctx context.Context) bool {
	e.authOnce.Do(func() {
		cmd := e.command(ctx, []string{"auth", "token", "--hostname", "github.com"}, "")
		// osexec's helpers output stdout to the terminal, so replace stdout.
		stdout := &bytes.Buffer{}
		cmd.Stdout = stdout
		cmd.Stderr = io.Discard
		if _, err := e.executor.Exec(cmd); err != nil {
			e.ghTokenKey = ghKeyAnonymous
			return
		}
		e.authenticated = true
		token := strings.TrimSpace(stdout.String())
		if token == "" {
			// It's unexpected, but the access token can't be compared.
			e.ghTokenKey = "gh:unknown"
			return
		}
		e.ghTokenKey = hashToken(token)
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
// The output isn't written to the terminal here. Verify writes it after all attempts.
// token is passed to GitHub CLI as GH_TOKEN. It's ignored if it's empty.
func (e *ExecutorImpl) run(ctx context.Context, args []string, token string) (string, error) {
	cmd := e.command(ctx, args, token)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out
	_, err := e.executor.Exec(cmd)
	return out.String(), err
}
