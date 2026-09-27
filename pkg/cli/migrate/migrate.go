// Package migrate implements the aqua migrate commands, which bring a configuration to the
// shape a newer aqua wants.
package migrate

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

// New creates the "aqua migrate" command.
//
// One subcommand per migration, rather than one command with a flag per migration. Each
// carries its own decision -- whether to keep aqua-checksums.json for whoever hasn't
// upgraded is not a question the others ask -- and a flag for it belongs where it is asked
// rather than on a command line where most of them mean nothing.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Bring the configuration to the shape a newer aqua wants",
		Commands: []*cli.Command{
			newChecksum(r, globalArgs),
		},
	}
}
