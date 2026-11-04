package ghattestation

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// noToken means gh is run without GH_TOKEN.
const noToken = "<none>"

type fakeExecutor struct {
	authenticated bool
	// ghToken is output by gh auth token.
	ghToken   string
	authCount int
	// authFailures is the number of the first gh auth token calls which fail.
	authFailures int
	// outputs are returned by gh attestation verify in order.
	// An empty string means success.
	outputs []string
	// tokens are GH_TOKEN passed to each gh attestation verify.
	tokens []string
}

func (f *fakeExecutor) Exec(cmd *exec.Cmd) (int, error) {
	if slices.Equal(cmd.Args[1:3], []string{"auth", "token"}) {
		f.authCount++
		if f.authenticated && f.authCount > f.authFailures {
			fmt.Fprintln(cmd.Stdout, f.ghToken)
			return 0, nil
		}
		return 1, errors.New("exit status 1")
	}
	out, code, err := f.verify(cmd)
	fmt.Fprint(cmd.Stdout, out)
	return code, err
}

// verify returns the output of gh attestation verify.
func (f *fakeExecutor) verify(cmd *exec.Cmd) (string, int, error) {
	token := noToken
	for _, env := range cmd.Env {
		if t, ok := strings.CutPrefix(env, "GH_TOKEN="); ok {
			token = t
		}
	}
	f.tokens = append(f.tokens, token)
	if len(f.outputs) == 0 {
		return outSuccess, 0, nil
	}
	out := f.outputs[0]
	f.outputs = f.outputs[1:]
	if out == "" {
		return outSuccess, 0, nil
	}
	return out, 1, errors.New("exit status 1")
}

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

// tokenSourceFunc returns newTokenSource returning ts.
// nil ts means the credential isn't enabled.
func tokenSourceFunc(ts *countTokenSource, err error, count *int) newTokenSource {
	return func(_ *slog.Logger) (oauth2.TokenSource, error) {
		*count++
		if err != nil {
			return nil, err
		}
		if ts == nil {
			return nil, nil
		}
		return ts, nil
	}
}

const (
	outInvalidToken = "Error: HTTP 401: Bad credentials (https://api.github.com/repos/foo/bar/attestations/sha256:xxx)" //nolint:gosec // This is an output of GitHub CLI, not a credential.
	outNotFound     = "Error: HTTP 404: Not Found (https://api.github.com/repos/foo/bar/attestations/sha256:xxx)"
	outNoAtt        = "Error: no attestations found"
	outVerification = `Error: verifying with issuer "sigstore.dev"`
	outRateLimit    = "Error: HTTP 403: API rate limit exceeded for 192.0.2.1. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.) (https://api.github.com/repos/foo/bar/attestations/sha256:xxx)"
	outSAML         = "Error: HTTP 403: Resource protected by organization SAML enforcement. You must grant your Personal Access token access to this organization. (https://api.github.com/repos/foo/bar/attestations/sha256:xxx)"
	outTemporary    = "Error: HTTP 502: Bad Gateway (https://api.github.com/repos/foo/bar/attestations/sha256:xxx)"
	outAuth         = "To get started with GitHub CLI, please run:  gh auth login"
	outUnknown      = "unknown flag: --foo\n"
	outSuccess      = "Verification succeeded!\n"
)

func TestClassify(t *testing.T) {
	t.Parallel()
	data := []struct {
		out string
		exp failure
	}{
		{outInvalidToken, failureInvalidToken},
		{outNotFound, failureNotFound},
		{outNoAtt, failureNoAttestation},
		{outVerification, failureVerification},
		{"Error: expected SourceRepositoryURI to be https://github.com/foo/bar, got https://github.com/foo/baz", failureVerification},
		{outRateLimit, failureRateLimit},
		{"Error: HTTP 429: You have exceeded a secondary rate limit. (https://api.github.com/...)", failureRateLimit},
		{"Error: HTTP 403: You have exceeded a secondary rate limit. (https://api.github.com/...)", failureRateLimit},
		{outSAML, failurePermission},
		{outTemporary, failureTemporary},
		{"dial tcp: lookup api.github.com: no such host", failureTemporary},
		{outAuth, failureAuth},
		{"Error: HTTP 429: Too Many Requests (https://api.github.com/...)", failureRateLimit},
		{"unknown flag: --foo", failureUnknown},
		{"Error: context deadline exceeded", failureTemporary},
		{"Error: attestation bundle with URL https://example.com/bundle returned status code 503", failureTemporary},
		{"Error: failed to fetch bundle with URL: request to fetch bundle from URL failed: dial tcp", failureTemporary},
		{`Error: Get "https://api.github.com/...": net/http: request canceled (Client.Timeout exceeded while awaiting headers)`, failureTemporary},
		{"Error: http2: server sent GOAWAY and closed the connection", failureTemporary},
		{"Error: Get \"https://api.github.com/...\": EOF\n", failureTemporary},
		{"Error: unexpected EOF", failureTemporary},
		{"Error: error creating Sigstore verifier: no valid Sigstore verifiers could be initialized", failureTemporary},
		{"Error: failed to choose verifier based on provided bundle issuer: public good verifier is not available (initialization may have failed)", failureTemporary},
		{"Error: failed to create TUF client: connection failed", failureTemporary},
		{"Error: the EOF marker is invalid", failureUnknown},
		{outInvalidToken + "\nTry authenticating with:  gh auth login", failureInvalidToken},
	}
	for _, d := range data {
		if got := classify(d.out); got != d.exp {
			t.Errorf("classify(%q) = %s, wanted %s", d.out, got, d.exp)
		}
	}
}

type verifyCall struct {
	private bool
	// stderr is the expected output written to stderr. It's checked only if it isn't empty.
	stderr  string
	outputs []string
	wantErr bool
	// tokens are the expected GH_TOKEN passed to each gh attestation verify.
	tokens []string
}

type verifyTestCase struct {
	name          string
	authenticated bool
	// ghToken is GitHub CLI's own access token. It's "gh" if it's empty.
	ghToken    string
	aquaToken  string
	keyring    *countTokenSource
	keyringErr error
	ghtkn      *countTokenSource
	// calls are Verify calls run in order with the same executor.
	calls []verifyCall
}

func verifyTestCases() []*verifyTestCase { //nolint:funlen
	return []*verifyTestCase{
		{
			name:          "authenticated: gh's own credential is used first",
			authenticated: true,
			aquaToken:     "aqua",
			ghtkn:         &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{tokens: []string{noToken}, stderr: outSuccess},
			},
		},
		{
			name:          "authenticated: invalid gh's own credential falls back to AQUA_GITHUB_TOKEN",
			authenticated: true,
			aquaToken:     "aqua",
			calls: []verifyCall{
				{outputs: []string{outInvalidToken}, tokens: []string{noToken, "aqua"}},
				// gh's own credential isn't used anymore.
				{tokens: []string{"aqua"}},
			},
		},
		{
			name:      "not authenticated: AQUA_GITHUB_TOKEN > keyring > ghtkn > none",
			aquaToken: "aqua",
			keyring:   &countTokenSource{token: "keyring"},
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{
					private: true,
					outputs: []string{outNotFound, outNotFound, outNotFound, outNotFound},
					wantErr: true,
					tokens:  []string{"aqua", "keyring", "ghtkn", noToken},
				},
			},
		},
		{
			name:  "not authenticated without any token: gh is run as is",
			calls: []verifyCall{{tokens: []string{noToken}}},
		},
		{
			name:      "404 for a public package fails immediately",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outNotFound}, wantErr: true, tokens: []string{"aqua"}},
			},
		},
		{
			name:      "404 for a private package tries the next token",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{private: true, outputs: []string{outNotFound}, tokens: []string{"aqua", "ghtkn"}},
				// 404 doesn't disable the token.
				{tokens: []string{"aqua"}},
			},
		},
		{
			name:      "no attestations fails immediately",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{private: true, outputs: []string{outNoAtt}, wantErr: true, tokens: []string{"aqua"}},
			},
		},
		{
			name:      "verification failure fails immediately",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outVerification}, wantErr: true, tokens: []string{"aqua"}},
			},
		},
		{
			name:      "rate limit disables the token",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outRateLimit}, tokens: []string{"aqua", "ghtkn"}},
				{tokens: []string{"ghtkn"}},
			},
		},
		{
			name:      "permission error doesn't disable the token",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outSAML}, tokens: []string{"aqua", "ghtkn"}},
				{tokens: []string{"aqua"}},
			},
		},
		{
			name:      "temporary errors are retried with the same token",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outTemporary, outTemporary}, tokens: []string{"aqua", "aqua", "aqua"}},
			},
		},
		{
			name:      "retries are limited, and then the next token is tried once each",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{
					outputs: []string{outTemporary, outTemporary, outTemporary, outTemporary, outTemporary, outTemporary, outTemporary},
					wantErr: true,
					tokens:  []string{"aqua", "aqua", "aqua", "aqua", "aqua", "ghtkn", noToken},
				},
			},
		},
		{
			name:      "the next token is tried after the retries are exhausted",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{
					outputs: []string{outTemporary, outTemporary, outTemporary, outTemporary, outTemporary},
					tokens:  []string{"aqua", "aqua", "aqua", "aqua", "aqua", "ghtkn"},
				},
			},
		},
		{
			name:      "the same token isn't tried twice",
			aquaToken: "same",
			keyring:   &countTokenSource{token: "same"},
			calls: []verifyCall{
				{private: true, outputs: []string{outNotFound, outNotFound}, wantErr: true, tokens: []string{"same", noToken}},
			},
		},
		{
			name:          "spaces around tokens are trimmed",
			authenticated: true,
			ghToken:       "same",
			aquaToken:     "same\n",
			ghtkn:         &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outSAML}, tokens: []string{noToken, "ghtkn"}},
			},
		},
		{
			name:          "the same token as gh's own access token is skipped",
			authenticated: true,
			ghToken:       "same",
			aquaToken:     "same",
			ghtkn:         &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outSAML}, tokens: []string{noToken, "ghtkn"}},
			},
		},
		{
			name:          "an invalid token is disabled regardless of its source",
			authenticated: true,
			ghToken:       "same",
			aquaToken:     "same",
			ghtkn:         &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outInvalidToken}, tokens: []string{noToken, "ghtkn"}},
				// Neither gh's own access token nor AQUA_GITHUB_TOKEN is used.
				{tokens: []string{"ghtkn"}},
			},
		},
		{
			name:      "a rate limited token is disabled regardless of its source",
			aquaToken: "same",
			keyring:   &countTokenSource{token: "same"},
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outRateLimit}, tokens: []string{"same", "ghtkn"}},
				{tokens: []string{"ghtkn"}},
			},
		},
		{
			name:  "running gh without any token is disabled by the rate limit",
			ghtkn: &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outRateLimit, outRateLimit}, wantErr: true, tokens: []string{"ghtkn", noToken}},
				{wantErr: true},
			},
		},
		{
			name:      "an unknown failure fails immediately",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outUnknown}, wantErr: true, tokens: []string{"aqua"}, stderr: outUnknown},
			},
		},
		{
			name:      "only the output of the successful attempt is written",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outRateLimit}, tokens: []string{"aqua", "ghtkn"}, stderr: outSuccess},
			},
		},
		{
			name:      "all outputs are written if the verification fails",
			aquaToken: "aqua",
			ghtkn:     &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{outputs: []string{outRateLimit, outVerification}, wantErr: true, tokens: []string{"aqua", "ghtkn"}, stderr: outRateLimit + outVerification},
			},
		},
		{
			name:       "the failure of the keyring is warned only once",
			keyringErr: errors.New("keyring is broken"),
			ghtkn:      &countTokenSource{token: "ghtkn"},
			calls: []verifyCall{
				{tokens: []string{"ghtkn"}},
				{tokens: []string{"ghtkn"}},
			},
		},
		{
			name:      "the authentication error doesn't skip the verification",
			aquaToken: "aqua",
			calls: []verifyCall{
				{outputs: []string{outInvalidToken, outAuth}, wantErr: true, tokens: []string{"aqua", noToken}},
			},
		},
		{
			name:          "all tokens are disabled",
			authenticated: true,
			calls: []verifyCall{
				{outputs: []string{outInvalidToken}, wantErr: true, tokens: []string{noToken}},
				{wantErr: true},
			},
		},
	}
}

func TestExecutorImpl_Verify(t *testing.T) {
	t.Parallel()
	for _, tt := range verifyTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testVerify(t, tt)
		})
	}
}

func checkVerifyCall(t *testing.T, i int, call *verifyCall, err error, tokens []string, stderr string) {
	t.Helper()
	if call.wantErr && err == nil {
		t.Fatalf("call %d: error must be returned", i)
	}
	if !call.wantErr && err != nil {
		t.Fatalf("call %d: unexpected error: %v", i, err)
	}
	if !slices.Equal(tokens, call.tokens) {
		t.Fatalf("call %d: tokens = %v, wanted %v", i, tokens, call.tokens)
	}
	if call.stderr != "" && stderr != call.stderr {
		t.Fatalf("call %d: stderr = %q, wanted %q", i, stderr, call.stderr)
	}
}

func testVerify(t *testing.T, tt *verifyTestCase) {
	t.Helper()
	keyringCreated := 0
	ghtknCreated := 0
	exe := &fakeExecutor{authenticated: tt.authenticated, ghToken: cmp.Or(tt.ghToken, "gh")}
	envs := map[string]string{"AQUA_GITHUB_TOKEN": tt.aquaToken}
	e := newExecutor(exe, "gh", func(k string) string { return envs[k] }, func() []string { return nil },
		tokenSourceFunc(tt.keyring, tt.keyringErr, &keyringCreated),
		tokenSourceFunc(tt.ghtkn, nil, &ghtknCreated))
	e.wait = func(context.Context, *slog.Logger, int) error { return nil }
	stderr := &strings.Builder{}
	e.stderr = stderr
	for i, call := range tt.calls {
		exe.outputs = call.outputs
		exe.tokens = nil
		stderr.Reset()
		err := e.Verify(t.Context(), slog.New(slog.DiscardHandler), &ParamVerify{
			ArtifactPath: "foo",
			Repository:   "foo/bar",
			Private:      call.private,
		})
		checkVerifyCall(t, i, &call, err, exe.tokens, stderr.String())
	}
	if exe.authCount > 1 {
		t.Fatalf("gh auth token must be run at most once, but it was run %d times", exe.authCount)
	}
	// Token sources are created at most once even if they fail.
	if keyringCreated > 1 {
		t.Fatalf("the keyring token source was created %d times", keyringCreated)
	}
	if ghtknCreated > 1 {
		t.Fatalf("the ghtkn token source was created %d times", ghtknCreated)
	}
}

func TestFailure_String(t *testing.T) {
	t.Parallel()
	data := map[failure]string{
		failureNone:          "none",
		failureTemporary:     "temporary",
		failureVerification:  "verification",
		failureNoAttestation: "no_attestation",
		failureNotFound:      "not_found",
		failureInvalidToken:  "invalid_token",
		failureRateLimit:     "rate_limit",
		failurePermission:    "permission",
		failureUnknown:       "unknown",
		failureAuth:          "auth",
		failureAuth + 1:      "invalid",
	}
	for f, exp := range data {
		if got := f.String(); got != exp {
			t.Errorf("failure(%d).String() = %s, wanted %s", f, got, exp)
		}
	}
}

func TestExecutorImpl_Verify_canceled(t *testing.T) {
	t.Parallel()
	// gh auth token fails because of the cancellation at first.
	exe := &fakeExecutor{authenticated: true, ghToken: "gh", authFailures: 1}
	e := newExecutor(exe, "gh", func(string) string { return "" }, func() []string { return nil },
		tokenSourceFunc(nil, nil, new(int)), tokenSourceFunc(nil, nil, new(int)))
	e.wait = func(context.Context, *slog.Logger, int) error { return nil }
	e.stderr = &strings.Builder{}
	logger := slog.New(slog.DiscardHandler)
	param := &ParamVerify{ArtifactPath: "foo", Repository: "foo/bar"}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	exe.outputs = []string{outUnknown}
	err := e.Verify(ctx, logger, param)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("the context error must be returned, but got %v", err)
	}
	if errors.Is(err, errVerify) {
		t.Fatal("the cancellation must not be reported as the verification failure")
	}

	// The result of gh auth token isn't cached if the context is canceled.
	exe.outputs = nil
	exe.tokens = nil
	if err := e.Verify(t.Context(), logger, param); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exe.authCount != 2 {
		t.Fatalf("gh auth token must be run again, but it was run %d times", exe.authCount)
	}
	// gh's own credential is used first, because gh is authenticated.
	if !slices.Equal(exe.tokens, []string{noToken}) {
		t.Fatalf("tokens = %v, wanted [%s]", exe.tokens, noToken)
	}
	// The successful result is cached.
	if err := e.Verify(t.Context(), logger, param); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exe.authCount != 2 {
		t.Fatalf("the result of gh auth token must be cached, but it was run %d times", exe.authCount)
	}
}
