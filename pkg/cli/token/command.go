// Package token implements the aqua token command for managing GitHub tokens.
// The token command provides functionality to store and manage GitHub tokens
// securely using the system keyring.
package token

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/keyring"
	ghTokenCLI "github.com/suzuki-shunsuke/urfave-cli-v3-util/keyring/ghtoken/cli"
	"github.com/urfave/cli/v3"
)

// New creates and returns a new CLI command for token management.
// It integrates with the GitHub token CLI utility to provide secure
// token storage and retrieval functionality using the system keyring.
func New(r *util.Param) *cli.Command {
	return ghTokenCLI.New(r.LogE, keyring.KeyService)
}
