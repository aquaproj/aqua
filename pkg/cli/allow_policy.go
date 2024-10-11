package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type policyAllowCommand struct {
	r *Runner
}

func newPolicyAllow(r *Runner) *cli.Command {
	i := &policyAllowCommand{
		r: r,
	}
	return &cli.Command{
		Action: i.action,
		Name:   "allow",
		Usage:  "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua policy allow [<policy file path>]
`,
	}
}

func (pa *policyAllowCommand) action(c *cli.Context) error {
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
	if err := setParam(c, pa.r.LogE, "allow-policy", param, pa.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(c.Context, param)
	return ctrl.Allow(pa.r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
