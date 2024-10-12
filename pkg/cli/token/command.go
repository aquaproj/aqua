package token

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v2"
)

func New(r *util.Param) *cli.Command {
	return &cli.Command{
		Name:  "token",
		Usage: "Manage a GitHub Access Token in keyring",
		Subcommands: []*cli.Command{
			newSet(r),
		},
	}
}
