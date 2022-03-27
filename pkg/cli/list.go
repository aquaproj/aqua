package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List packages in Registries",
		Action: runner.listAction,
		Description: `Output the list of packages in registries.
The output format is <registry name>,<package name>

e.g.
$ aqua list
standard,99designs/aws-vault
standard,abiosoft/colima
standard,abs-lang/abs
...
`,
	}
}

func (runner *Runner) listAction(c *cli.Context) error {
	param := &config.Param{}
	if err := runner.setCLIArg(c, param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}

	ctrl, err := controller.NewController(c.Context, param.AQUAVersion, param)
	if err != nil {
		return fmt.Errorf("initialize a controller: %w", err)
	}

	return ctrl.List(c.Context, param, c.Args().Slice()) //nolint:wrapcheck
}
