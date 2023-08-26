package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newDenyPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "deny",
		Usage: "Deny a policy file",
		Description: `Deny a policy file
e.g.
$ aqua policy deny [<policy file path>]
`,
		Action: r.denyPolicyAction,
	}
}

func (r *Runner) denyPolicyAction(c *cli.Context) error {
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
	if err := r.setParam(c, "deny-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(c.Context, param)
	return ctrl.Deny(c.Context, r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
