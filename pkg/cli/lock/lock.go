// Package lock implements the aqua lock commands, which manage aqua-lock.json.
package lock

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

// New creates the "aqua lock" command.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	return &cli.Command{
		Name:  "lock",
		Usage: "Manage the lock file",
		Commands: []*cli.Command{
			newUpdate(r, globalArgs),
		},
	}
}
