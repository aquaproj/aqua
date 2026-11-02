package ghattestation

import (
	"errors"
	"log/slog"
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
	name   string
	envs   map[string]string
	ts     *countTokenSource // nil means neither the keyring nor ghtkn is enabled
	tsErr  error
	exp    string
	called bool
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
			name:   "keyring or ghtkn",
			ts:     &countTokenSource{token: "yyy"},
			exp:    "GH_TOKEN=yyy",
			called: true,
		},
		{
			name: "neither keyring nor ghtkn is enabled",
		},
		{
			name:   "failed to get a token",
			ts:     &countTokenSource{err: errToken},
			called: true,
		},
		{
			name:  "failed to create a token source",
			tsErr: errToken,
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
			e := &ExecutorImpl{
				getEnv: func(k string) string {
					return d.envs[k]
				},
				newTokenSource: func(*slog.Logger) (oauth2.TokenSource, error) {
					created++
					if d.tsErr != nil || d.ts == nil {
						return nil, d.tsErr
					}
					return d.ts, nil
				},
			}
			// call twice to check the token source is created only once
			for range 2 {
				if env := e.tokenEnv(logger); env != d.exp {
					t.Fatalf("wanted %q, got %q", d.exp, env)
				}
			}
			if created > 1 {
				t.Fatalf("token source must be created only once, but created %d times", created)
			}
			if d.ts == nil {
				return
			}
			if d.called != (d.ts.count > 0) {
				t.Fatalf("token source called: wanted %v, got %d times", d.called, d.ts.count)
			}
		})
	}
}
