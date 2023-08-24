package cli

import (
	"github.com/urfave/cli/v2"
)

func (r *Runner) newPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Subcommands: []*cli.Command{
			r.newAllowPolicyCommand(),
			r.newDenyPolicyCommand(),
			r.newPolicyInitCommand(),
		},
	}
}
