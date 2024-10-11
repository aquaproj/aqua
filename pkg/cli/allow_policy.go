package cli //nolint:dupl

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type policyAllowCommand struct {
	logE    *logrus.Entry
	ldFlags *LDFlags
}

func (pa *policyAllowCommand) command() *cli.Command {
	return &cli.Command{
		Name:  "allow",
		Usage: "Allow a policy file",
		Description: `Allow a policy file
e.g.
$ aqua policy allow [<policy file path>]
`,
		Action: pa.action,
		Flags:  []cli.Flag{},
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
	if err := setParam(c, pa.logE, "allow-policy", param, pa.ldFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeAllowPolicyCommandController(c.Context, param)
	return ctrl.Allow(pa.logE, param, c.Args().First()) //nolint:wrapcheck
}
