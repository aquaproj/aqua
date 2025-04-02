package policy //nolint:dupl

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type policyDenyCommand struct {
	r *util.Param
}

func newPolicyDeny(r *util.Param) *cli.Command {
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

func (pd *policyDenyCommand) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, pd.r.LogE, "deny-policy", param, pd.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(ctx, param)
	return ctrl.Deny(pd.r.LogE, param, cmd.Args().First()) //nolint:wrapcheck
}
