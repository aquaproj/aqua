package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newSetPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `Create a policy file if it doesn't exist
e.g.
$ aqua policy set # create "aqua-policy.yaml"
$ aqua policy set foo.yaml # create foo.yaml`,
		Action: runner.initPolicyAction,
	}
}
