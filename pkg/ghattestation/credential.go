package ghattestation

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"golang.org/x/oauth2"
)

// candidate is a GitHub access token to try.
type candidate struct {
	// name is the source of the access token. It's used for logging.
	name string
	// token is passed to GitHub CLI as GH_TOKEN.
	// It's empty for GitHub CLI's own credential, which GitHub CLI reads by itself.
	token string
	// key identifies the access token without holding the raw access token.
	// It's the SHA-256 hash of the access token, so the same access token from different sources has the same key.
	key string
}

// hashToken returns the SHA-256 hash of an access token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// credential is a source of a GitHub access token passed to GitHub CLI as GH_TOKEN.
// Its state is shared among packages installed in parallel.
type credential struct {
	name string
	// newTokenSource creates a token source. It returns nil if the credential isn't enabled.
	newTokenSource newTokenSource

	mutex       sync.Mutex
	tokenSource oauth2.TokenSource
	// unavailable is true if the credential isn't enabled or fails to get an access token.
	unavailable bool
}

// candidate returns a candidate of the credential.
// ok is false if the credential isn't available.
// The token source is created lazily, because getting an access token from ghtkn may require the authentication.
// If it fails to get an access token, it outputs a warning and makes the credential unavailable for the rest of the process,
// so that it doesn't try to get an access token and output the same warning for every package.
func (c *credential) candidate(logger *slog.Logger) (cand candidate, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.unavailable {
		return candidate{}, false
	}
	if c.tokenSource == nil {
		ts, err := c.newTokenSource(logger)
		if err != nil {
			slogerr.WithError(logger, err).Warn("create a token source for GitHub CLI", "credential", c.name)
		}
		if ts == nil {
			c.unavailable = true
			return candidate{}, false
		}
		c.tokenSource = ts
	}
	// Token sources cache the access token, so this doesn't call any API every time.
	t, err := c.tokenSource.Token()
	if err != nil {
		c.unavailable = true
		slogerr.WithError(logger, err).Warn("get a GitHub access token for GitHub CLI", "credential", c.name)
		return candidate{}, false
	}
	if t.AccessToken == "" {
		c.unavailable = true
		return candidate{}, false
	}
	return candidate{name: c.name, token: t.AccessToken, key: hashToken(t.AccessToken)}, true
}
