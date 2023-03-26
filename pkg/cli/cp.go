package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newCpCommand() *cli.Command {
	return &cli.Command{
		Name:      "cp",
		Usage:     "Copy executable files in a directory",
		ArgsUsage: `<command name> [<command name> ...]`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "o",
				Value: "dist",
				Usage: "destination directory",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "install all aqua configuration packages",
			},
			&cli.StringFlag{
				Name:    "tags",
				Aliases: []string{"t"},
				Usage:   "filter installed packages with tags",
			},
			&cli.StringFlag{
				Name:  "exclude-tags",
				Usage: "exclude installed packages with tags",
			},
		},
		Description: `Copy executable files in a directory.

e.g.
$ aqua cp gh
$ ls dist
gh

You can specify the target directory by -o option.

$ aqua cp -o ~/bin terraform hugo

If you don't specify commands, all commands are copied.

$ aqua cp

You can also copy global configuration files' commands with "-a" option.

$ aqua cp -a

You can filter copied commands with package tags.

e.g.
$ aqua cp -t foo # Copy only packages having a tag "foo"
$ aqua cp --exclude-tags foo # Copy only packages not having a tag "foo"
`,
		Action: runner.cpAction,
	}
}

func (runner *Runner) cpAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "cp", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	param.SkipLink = true
	ctrl := controller.InitializeCopyCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	if err := ctrl.Copy(c.Context, runner.LogE, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
