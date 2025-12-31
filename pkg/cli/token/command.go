// Package token implements the aqua token command for managing GitHub tokens.
// The token command provides functionality to store and manage GitHub tokens
// securely using the system keyring.
package token

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/keyring"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/keyring/ghtoken"
	"github.com/urfave/cli/v3"
)

// New creates a new token command for the CLI.
// It initializes a GitHub token management command using the system keyring
// for secure credential storage and retrieval.
// Returns a pointer to the configured CLI command for token operations.
func New(param *util.Param) *cli.Command {
	return ghtoken.Command(ghtoken.NewActor(param.Logger.Logger, keyring.KeyService))
}
