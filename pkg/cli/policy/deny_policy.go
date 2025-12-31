package policy //nolint:dupl

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// policyDenyArgs holds command-line arguments for the policy deny command.
type policyDenyArgs struct {
	*cliargs.GlobalArgs

	PolicyPath string
}

// policyDenyCommand holds the parameters and configuration for the policy deny command.
type policyDenyCommand struct {
	r *util.Param
}

// newPolicyDeny creates and returns a new CLI command for denying policy files.
// The returned command marks a policy file as denied, preventing packages
// from being installed according to that policy.
func newPolicyDeny(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &policyDenyArgs{
		GlobalArgs: globalArgs,
	}
	i := &policyDenyCommand{
		r: r,
	}
	return &cli.Command{
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Name:  "deny",
		Usage: "Deny a policy file",
		Description: `Deny a policy file
e.g.
$ aqua policy deny [<policy file path>]
`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "policy_path",
				Destination: &args.PolicyPath,
			},
		},
	}
}

// action implements the main logic for the policy deny command.
// It initializes the deny policy controller and marks the specified
// policy file as denied based on the provided file path.
func (pd *policyDenyCommand) action(ctx context.Context, args *policyDenyArgs) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, pd.r.Logger, param, pd.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(ctx, param)
	return ctrl.Deny(pd.r.Logger.Logger, param, args.PolicyPath) //nolint:wrapcheck
}
