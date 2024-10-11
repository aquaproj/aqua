package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type policyDenyCommand struct {
	r *Runner
}

func newPolicyDeny(r *Runner) *cli.Command {
	i := &policyDenyCommand{
		r: r,
	}
	return &cli.Command{
		Action: i.action,
		Name:   "deny",
		Usage:  "Deny a policy file",
		Description: `Deny a policy file
e.g.
$ aqua policy deny [<policy file path>]
`,
	}
}

func (pd *policyDenyCommand) action(c *cli.Context) error {
	tracer, err := startTrace(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := startCPUProfile(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	param := &config.Param{}
	if err := setParam(c, pd.r.LogE, "deny-policy", param, pd.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(c.Context, param)
	return ctrl.Deny(pd.r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
