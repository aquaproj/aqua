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

// badCredentialsExecutor fails with "Bad credentials" if GH_TOKEN is passed.
type badCredentialsExecutor struct {
	envs [][]string
}

func (b *badCredentialsExecutor) ExecStderrAndGetCombinedOutput(cmd *exec.Cmd) (string, int, error) {
	b.envs = append(b.envs, cmd.Env)
	if slices.ContainsFunc(cmd.Env, func(s string) bool { return strings.HasPrefix(s, "GH_TOKEN=") }) {
		return "HTTP 401: Bad credentials (https://api.github.com/repos/foo/bar/attestations)", 1, errors.New("exit status 1")
	}
	return "", 0, nil
}

func TestExecutorImpl_exec_badCredentials(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	created := 0
	ts := &countTokenSource{token: "revoked"}
	e := newTestExecutor(nil, ts, nil, &created)
	exe := &badCredentialsExecutor{}
	e.executor = exe
	e.exePath = "gh"
	for range 2 {
		if err := e.exec(t.Context(), logger, []string{"attestation", "verify"}); err != nil {
			t.Fatal(err)
		}
	}
	// 1st call: with the token (fails), then without the token (succeeds)
	// 2nd call: without the token, because the token is disabled
	if len(exe.envs) != 3 {
		t.Fatalf("gh must be executed 3 times, but executed %d times", len(exe.envs))
	}
	if ts.count != 1 {
		t.Fatalf("token source must be called once, but called %d times", ts.count)
	}
}
