package cli

import (
	"github.com/urfave/cli/v2"
)

type initPolicyCommand struct {
	r *Runner
}

func newInitPolicy(r *Runner) *cli.Command {
	cmd := newPolicyInit(r)
	return &cli.Command{
		Name:      "init-policy",
		Usage:     "[Deprecated] Create a policy file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua-policy.yaml">]`,
		Description: `[Deprecated] Create a policy file if it doesn't exist

Please use "aqua policy init" command instead.

e.g.
$ aqua init-policy # create "aqua-policy.yaml"
$ aqua init-policy foo.yaml # create foo.yaml`,
		Action: cmd.Action,
	}
}
