package cli

import (
	"fmt"
	"net/http"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newInstallCommand() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "Install tools",
		Description: `Install tools according to the configuration files.

e.g.
$ clivm i

If you want to create only symbolic links and want to skip downloading package, please set "-l" option.

$ clivm i -l

By default clivm doesn't install packages in the global configuration.
If you want to install packages in the global configuration too,
please set "-a" option.

$ clivm i -a
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
				Usage:   "install all clivm configuration packages",
			},
		},
	}
}

func (runner *Runner) installAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "install", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInstallCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	return ctrl.Install(c.Context, param, runner.LogE) //nolint:wrapcheck
}
