package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/aquaproj/aqua/pkg/log"
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
	logE := log.New(param.AQUAVersion)
	log.SetLevel(param.LogLevel, logE)
	ctrl := controller.InitializeInitCommandController(c.Context, param.AQUAVersion, param)
	return ctrl.Init(c.Context, c.Args().First(), logE) //nolint:wrapcheck
}
