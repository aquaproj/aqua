package token

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/keyring"
	ghTokenCLI "github.com/suzuki-shunsuke/urfave-cli-v3-util/keyring/ghtoken/cli"
	"github.com/urfave/cli/v3"
)

func New(r *util.Param) *cli.Command {
	return ghTokenCLI.New(r.LogE, keyring.KeyService)
}
