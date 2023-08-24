package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newInstallCommand() *cli.Command {
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

You can filter installed packages with package tags.

e.g.
$ aqua i -t foo # Install only packages having a tag "foo"
$ aqua i --exclude-tags foo # Install only packages not having a tag "foo"
`,
		Action: r.installAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "only-link",
				Aliases: []string{"l"},
				Usage:   "create links but skip downloading packages",
			},
			&cli.BoolFlag{
				Name:  "test",
				Usage: "This flag was deprecated and had no meaning from aqua v2.0.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/1691",
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
	}
}

func (r *Runner) installAction(c *cli.Context) error {
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
	if err := r.setParam(c, "install", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeInstallCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.Install(c.Context, r.LogE, param) //nolint:wrapcheck
}
