package ghattestation

import (
	"log/slog"
	"sync"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"golang.org/x/oauth2"
)

// credential is a source of a GitHub access token passed to GitHub CLI.
// Its state is shared among packages installed in parallel.
type credential struct {
	name string
	// newTokenSource creates a token source. It returns nil if the credential isn't enabled.
	// nil newTokenSource means GitHub CLI's own credential, which is used by passing no access token.
	newTokenSource newTokenSource

	mutex       sync.Mutex
	created     bool
	tokenSource oauth2.TokenSource
	// disabled is true if the credential is unavailable or has failed in a way that would fail for other packages too.
	disabled bool
}

// token returns an access token to pass to GitHub CLI.
// ok is false if the credential isn't available.
// token is empty for GitHub CLI's own credential.
// If it fails to get an access token, it outputs a warning and disables the credential for the rest of the process,
// so that it doesn't try to get an access token and output the same warning for every package.
func (c *credential) token(logger *slog.Logger) (token string, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.disabled {
		return "", false
	}
	if c.newTokenSource == nil {
		return "", true
	}
	if !c.created {
		c.created = true
		ts, err := c.newTokenSource(logger)
		if err != nil {
			c.disabled = true
			slogerr.WithError(logger, err).Warn("create a token source for GitHub CLI", "credential", c.name)
			return "", false
		}
		c.tokenSource = ts
	}
	if c.tokenSource == nil {
		return "", false
	}
	// Token sources cache the access token, so this doesn't call any API every time.
	t, err := c.tokenSource.Token()
	if err != nil {
		c.disabled = true
		slogerr.WithError(logger, err).Warn("get a GitHub access token for GitHub CLI", "credential", c.name)
		return "", false
	}
	if t.AccessToken == "" {
		return "", false
	}
	return t.AccessToken, true
}

// disable disables the credential for the rest of the process.
// It returns true if the credential has been enabled until now.
// The return value is used to output a warning only once even if gh is run concurrently.
func (c *credential) disable() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.disabled {
		return false
	}
	c.disabled = true
	return true
}
