package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Uninstall packages",
		ArgsUsage: `[<registry name>,]<package name> [...]`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "uninstall all packages",
			},
		},
		Description: `Uninstall packages.

e.g.
$ aqua rm --all
$ aqua rm cli/cli direnv/direnv

If you want to uninstall packages of non standard registry, you need to specify the registry name too.

e.g.
$ aqua rm foo,suzuki-shunsuke/foo

Limitation:
"http" and "go_install" packages can't be removed.
`,
		Action: r.removeAction,
	}
}

func (r *Runner) removeAction(c *cli.Context) error {
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
	if err := r.setParam(c, "remove", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	param.SkipLink = true
	ctrl := controller.InitializeRemoveCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	if err := ctrl.Remove(c.Context, r.LogE, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
