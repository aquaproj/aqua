package cli

import (
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newPolicyInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `Create a policy file if it doesn't exist
e.g.
$ aqua policy init # create "aqua-policy.yaml"
$ aqua policy init foo.yaml # create foo.yaml`,
		Action: runner.initPolicyAction,
	}
}
