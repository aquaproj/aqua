package github

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyService = "aquaproj.github.io"
	keyName    = "GITHUB_TOKEN"
)

func getTokenFromKeyring() (string, error) {
	s, err := keyring.Get(keyService, keyName)
	if err != nil {
		return "", fmt.Errorf("get a GitHub Access token from keyring: %w", err)
	}
	return s, nil
}

func SetTokenInKeyring(token string) error {
	if err := keyring.Set(keyService, keyName, token); err != nil {
		return fmt.Errorf("set a GitHub Access token in keyring: %w", err)
	}
	return nil
}
