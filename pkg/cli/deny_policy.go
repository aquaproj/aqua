package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type policyDenyCommand struct {
	logE    *logrus.Entry
	ldFlags *LDFlags
}

func (pd *policyDenyCommand) command() *cli.Command {
	return &cli.Command{
		Name:  "deny",
		Usage: "Deny a policy file",
		Description: `Deny a policy file
e.g.
$ aqua policy deny [<policy file path>]
`,
		Action: pd.action,
		Flags:  []cli.Flag{},
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
	if err := setParam(c, pd.logE, "deny-policy", param, pd.ldFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeDenyPolicyCommandController(c.Context, param)
	return ctrl.Deny(pd.logE, param, c.Args().First()) //nolint:wrapcheck
}
