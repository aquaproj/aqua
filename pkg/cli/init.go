package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a configuration file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua.yaml">]`,
		Description: `Create a configuration file if it doesn't exist
e.g.
$ aqua init # create "aqua.yaml"
$ aqua init foo.yaml # create foo.yaml`,
		Action: runner.initAction,
	}
}

func (runner *Runner) initAction(c *cli.Context) error {
	param := &config.Param{}
	if err := runner.setCLIArg(c, param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.NewController(c.Context, param.AQUAVersion, param)
	if err != nil {
		return fmt.Errorf("initialize a controller: %w", err)
	}
	return ctrl.Init(c.Context, c.Args().First()) //nolint:wrapcheck
}
