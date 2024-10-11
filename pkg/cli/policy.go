package cli

import (
	"github.com/urfave/cli/v2"
)

type policyCommand struct {
	r *Runner
}

func newPolicy(r *Runner) *cli.Command {
	return &cli.Command{
		Name:  "policy",
		Usage: "Manage Policy",
		Subcommands: []*cli.Command{
			newPolicyAllow(r),
			newPolicyDeny(r),
			newPolicyInit(r),
		},
	}
}
