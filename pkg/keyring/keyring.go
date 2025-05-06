// Package keyring provides a way to manage a GitHub access token using the system's keyring.
package keyring

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	keyService = "aquaproj.github.io"
	keyName    = "GITHUB_TOKEN"
)

func Enabled() bool {
	return os.Getenv("AQUA_KEYRING_ENABLED") == "true"
}

type TokenManager struct{}

func NewTokenManager() *TokenManager {
	return &TokenManager{}
}

func (tm *TokenManager) Get() (string, error) {
	s, err := keyring.Get(keyService, keyName)
	if err != nil {
		return "", fmt.Errorf("get a GitHub Access token from keyring: %w", err)
	}
	return s, nil
}

func (tm *TokenManager) Set(token string) error {
	if err := keyring.Set(keyService, keyName, token); err != nil {
		return fmt.Errorf("set a GitHub Access token in keyring: %w", err)
	}
	return nil
}

func (tm *TokenManager) Remove(logE *logrus.Entry) error {
	if err := keyring.Delete(keyService, keyName); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			logerr.WithError(logE, err).Warn("remove a GitHub Access token from keyring")
			return nil
		}
		return fmt.Errorf("delete a GitHub Access token from keyring: %w", err)
	}
	return nil
}

type TokenSource struct {
	token *oauth2.Token
	logE  *logrus.Entry
}

func NewTokenSource(logE *logrus.Entry) *TokenSource {
	return &TokenSource{
		logE: logE,
	}
}

func (ks *TokenSource) Token() (*oauth2.Token, error) {
	if ks.token != nil {
		return ks.token, nil
	}
	ks.logE.Debug("getting a GitHub Access toke from keyring")
	s, err := keyring.Get(keyService, keyName)
	if err != nil {
		return nil, fmt.Errorf("get a GitHub Access token from keyring: %w", err)
	}
	ks.logE.Debug("got a GitHub Access toke from keyring")
	ks.token = &oauth2.Token{
		AccessToken: s,
	}
	return ks.token, nil
}
