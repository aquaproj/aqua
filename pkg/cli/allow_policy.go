package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newAllowPolicyCommand() *cli.Command {
	return &cli.Command{
		Name:  "allow",
		Usage: "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua policy allow [<policy file path>]
`,
		Action: r.allowPolicyAction,
	}
}

func (r *Runner) allowPolicyAction(c *cli.Context) error {
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
	if err := r.setParam(c, "allow-policy", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(c.Context, param)
	return ctrl.Allow(c.Context, r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
