package ghattestation

import (
	"errors"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

type countTokenSource struct {
	token string
	err   error
	count int
}

func (c *countTokenSource) Token() (*oauth2.Token, error) {
	c.count++
	if c.err != nil {
		return nil, c.err
	}
	return &oauth2.Token{AccessToken: c.token}, nil
}

type tokenEnvTestCase struct {
	name      string
	envs      map[string]string
	ts        *countTokenSource // nil means neither the keyring nor ghtkn is enabled
	tsErr     error
	exp       string
	fromStore bool
	// count is the expected number of calls of the token source.
	count int
}

func tokenEnvTestCases() []*tokenEnvTestCase {
	errToken := errors.New("failed to get a token")
	return []*tokenEnvTestCase{
		{
			name: "GITHUB_TOKEN is set",
			envs: map[string]string{"GITHUB_TOKEN": "xxx"},
			ts:   &countTokenSource{token: "yyy"},
		},
		{
			name: "GH_TOKEN is set",
			envs: map[string]string{"GH_TOKEN": "xxx"},
			ts:   &countTokenSource{token: "yyy"},
		},
		{
			name: "AQUA_GITHUB_TOKEN is set",
			envs: map[string]string{"AQUA_GITHUB_TOKEN": "xxx"},
			ts:   &countTokenSource{token: "yyy"},
			exp:  "GITHUB_TOKEN=xxx",
		},
		{
			name:      "keyring or ghtkn",
			ts:        &countTokenSource{token: "yyy"},
			exp:       "GH_TOKEN=yyy",
			fromStore: true,
			count:     2,
		},
		{
			name: "neither keyring nor ghtkn is enabled",
		},
		{
			// the error is cached, so the token source is called only once
			name:  "failed to get a token",
			ts:    &countTokenSource{err: errToken},
			count: 1,
		},
		{
			name:  "failed to create a token source",
			tsErr: errToken,
		},
	}
}

func newTestExecutor(envs map[string]string, ts oauth2.TokenSource, tsErr error, created *int) *ExecutorImpl {
	return &ExecutorImpl{
		getEnv: func(k string) string {
			return envs[k]
		},
		environ: func() []string {
			return nil
		},
		newTokenSource: func(*slog.Logger) (oauth2.TokenSource, error) {
			*created++
			if tsErr != nil || ts == nil {
				return nil, tsErr
			}
			return ts, nil
		},
	}
}

func TestExecutorImpl_tokenEnv(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	for _, d := range tokenEnvTestCases() {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			created := 0
			var ts oauth2.TokenSource
			if d.ts != nil {
				ts = d.ts
			}
			e := newTestExecutor(d.envs, ts, d.tsErr, &created)
			// call twice to check the token source is created only once
			for range 2 {
				env, fromStore := e.tokenEnv(logger)
				if env != d.exp {
					t.Fatalf("wanted %q, got %q", d.exp, env)
				}
				if fromStore != d.fromStore {
					t.Fatalf("fromStore: wanted %v, got %v", d.fromStore, fromStore)
				}
			}
			if created > 1 {
				t.Fatalf("token source must be created only once, but created %d times", created)
			}
			if d.ts != nil && d.ts.count != d.count {
				t.Fatalf("token source must be called %d times, but called %d times", d.count, d.ts.count)
			}
		})
	}
}

type fakeResult struct {
	out string
	err error
}

// fakeExecutor returns withToken if GH_TOKEN is passed, otherwise withoutToken.
type fakeExecutor struct {
	withToken    fakeResult
	withoutToken fakeResult
	count        int
}

func (f *fakeExecutor) ExecStderrAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error) {
	f.count++
	r := f.withoutToken
	if slices.ContainsFunc(cmd.Env, func(s string) bool { return strings.HasPrefix(s, "GH_TOKEN=") }) {
		r = f.withToken
	}
	if r.err != nil {
		return r.out, 1, r.err
	}
	return r.out, 0, nil
}

type execTestCase struct {
	name         string
	withToken    fakeResult
	withoutToken fakeResult
	// errWithToken is true if the error with the token must be returned.
	errWithToken bool
	authError    bool
	disabled     bool
	count        int
}

func execTestCases() []*execTestCase {
	errExit := errors.New("exit status 1")
	authOut := "To get started with GitHub CLI, please run:  gh auth login"
	return []*execTestCase{
		{
			name:  "succeed with the token",
			count: 1,
		},
		{
			name:      "bad credentials, then succeed without the token",
			withToken: fakeResult{out: "HTTP 401: Bad credentials", err: errExit},
			disabled:  true,
			count:     2,
		},
		{
			name:      "no permission, then succeed without the token",
			withToken: fakeResult{out: "HTTP 403: Resource protected by organization SAML enforcement", err: errExit},
			disabled:  true,
			count:     2,
		},
		{
			name:         "bad credentials, then the authentication error without the token",
			withToken:    fakeResult{out: "HTTP 401: Bad credentials", err: errExit},
			withoutToken: fakeResult{out: authOut, err: errExit},
			authError:    true,
			disabled:     true,
			count:        2,
		},
		{
			// the verification failure must not be skipped as the authentication error
			name:         "verification failure with the token, then the authentication error without the token",
			withToken:    fakeResult{out: "verification failed", err: errExit},
			withoutToken: fakeResult{out: authOut, err: errors.New("exit status 4")},
			errWithToken: true,
			count:        2,
		},
	}
}

func TestExecutorImpl_exec(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	for _, d := range execTestCases() {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			created := 0
			e := newTestExecutor(nil, &countTokenSource{token: "xxx"}, nil, &created)
			exe := &fakeExecutor{withToken: d.withToken, withoutToken: d.withoutToken}
			e.executor = exe
			e.exePath = "gh"
			err := e.exec(t.Context(), logger, []string{"attestation", "verify"})
			checkExecError(t, d, err)
			if e.tokenDisabled != d.disabled {
				t.Fatalf("tokenDisabled: wanted %v, got %v", d.disabled, e.tokenDisabled)
			}
			if exe.count != d.count {
				t.Fatalf("gh must be executed %d times, but executed %d times", d.count, exe.count)
			}
		})
	}
}

func checkExecError(t *testing.T, d *execTestCase, err error) {
	t.Helper()
	ae := &AuthError{}
	isAuthError := errors.As(err, &ae)
	switch {
	case d.authError:
		if !isAuthError {
			t.Fatalf("AuthError must be returned, but got %v", err)
		}
	case d.errWithToken:
		if isAuthError || !errors.Is(err, d.withToken.err) {
			t.Fatalf("the error with the token must be returned, but got %v", err)
		}
	case err != nil:
		t.Fatal(err)
	}
}

func TestExecutorImpl_exec_disabledToken(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	created := 0
	ts := &countTokenSource{token: "revoked"}
	e := newTestExecutor(nil, ts, nil, &created)
	exe := &fakeExecutor{withToken: fakeResult{out: "HTTP 401: Bad credentials", err: errors.New("exit status 1")}}
	e.executor = exe
	e.exePath = "gh"
	for range 2 {
		if err := e.exec(t.Context(), logger, []string{"attestation", "verify"}); err != nil {
			t.Fatal(err)
		}
	}
	// 1st call: with the token (fails), then without the token (succeeds)
	// 2nd call: without the token, because the token is disabled
	if exe.count != 3 {
		t.Fatalf("gh must be executed 3 times, but executed %d times", exe.count)
	}
	if ts.count != 1 {
		t.Fatalf("token source must be called once, but called %d times", ts.count)
	}
}
