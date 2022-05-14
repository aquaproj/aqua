package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInstallCommand() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "Install tools",
		Description: `Install tools according to the configuration files.

e.g.
$ aqua i

If you want to create only symbolic links and want to skip downloading package, please set "-l" option.

$ aqua i -l

By default aqua doesn't install packages in the global configuration.
If you want to install packages in the global configuration too,
please set "-a" option.

$ aqua i -a
`,
		Action: runner.installAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "only-link",
				Aliases: []string{"l"},
				Usage:   "create links but skip downloading packages",
			},
			&cli.BoolFlag{
				Name:  "test",
				Usage: "test file.src after installing the package",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "install all aqua configuration packages",
			},
		},
	}
}

func (runner *Runner) installAction(c *cli.Context) error {
	param := &config.Param{}
	logE, err := runner.setParam(c, param)
	if err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInstallCommandController(c.Context, param)
	return ctrl.Install(c.Context, param, logE) //nolint:wrapcheck
}
