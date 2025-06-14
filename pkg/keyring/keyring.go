// Package keyring provides a way to manage a GitHub access token using the system's keyring.
package keyring

import (
	"os"
)

const (
	KeyService = "aquaproj.github.io"
)

func Enabled() bool {
	return os.Getenv("AQUA_KEYRING_ENABLED") == "true"
}
