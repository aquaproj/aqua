package policy //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

type policyAllowCommand struct {
	r *util.Param
}

func newPolicyAllow(r *util.Param) *cli.Command {
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
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(c, pa.r.LogE, "allow-policy", param, pa.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(c.Context, param)
	return ctrl.Allow(pa.r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
