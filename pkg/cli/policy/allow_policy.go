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

// policyAllowArgs holds command-line arguments for the policy allow command.
type policyAllowArgs struct {
	*cliargs.GlobalArgs

	PolicyPath []string
}

// policyAllowCommand holds the parameters and configuration for the policy allow command.
type policyAllowCommand struct {
	r *util.Param
}

// newPolicyAllow creates and returns a new CLI command for allowing policy files.
// The returned command marks a policy file as allowed, permitting packages
// to be installed according to that policy.
func newPolicyAllow(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &policyAllowArgs{
		GlobalArgs: globalArgs,
	}
	i := &policyAllowCommand{
		r: r,
	}
	return &cli.Command{
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Name:  "allow",
		Usage: "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua policy allow [<policy file path>]
`,
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "policy_path",
				Min:         0,
				Max:         1,
				Destination: &args.PolicyPath,
			},
		},
	}
}

// action implements the main logic for the policy allow command.
// It initializes the allow policy controller and marks the specified
// policy file as allowed based on the provided file path.
func (pa *policyAllowCommand) action(ctx context.Context, args *policyAllowArgs) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, pa.r.Logger, param, pa.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(ctx, param)
	policyPath := ""
	if len(args.PolicyPath) > 0 {
		policyPath = args.PolicyPath[0]
	}
	return ctrl.Allow(pa.r.Logger.Logger, param, policyPath) //nolint:wrapcheck
}
