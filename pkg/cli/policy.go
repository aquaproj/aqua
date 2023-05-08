package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Subcommands: []*cli.Command{
			runner.newAllowPolicyCommand(),
			runner.newDenyPolicyCommand(),
			runner.newPolicyInitCommand(),
		},
	}
}
