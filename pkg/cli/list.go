package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newListCommand() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List packages in Registries",
		Action: r.listAction,
		Description: `Output the list of packages in registries.
The output format is <registry name>,<package name>

e.g.
$ aqua list
standard,99designs/aws-vault
standard,abiosoft/colima
standard,abs-lang/abs
...

If the option -installed is set, the command lists only installed packages.

$ aqua list -installed
standard,golangci/golangci-lint,v1.56.2
standard,goreleaser/goreleaser,v1.24.0
...

By default, the command doesn't list global configuration packages.
If you want to list global configuration packages too, please set the option -a.

$ aqua list -installed -a
`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "installed",
				Usage: "List installed packages",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "List global configuration packages too",
			},
		},
	}
}

func (r *Runner) listAction(c *cli.Context) error {
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
	if err := r.setParam(c, "list", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeListCommandController(c.Context, param, http.DefaultClient, r.Runtime)
	return ctrl.List(c.Context, param, r.LogE) //nolint:wrapcheck
}
