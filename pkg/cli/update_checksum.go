package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newUpdateChecksumCommand() *cli.Command {
	return &cli.Command{
		Name:  "update-checksum",
		Usage: "Create or Update .aqua-checksums.json",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Create or Update all .aqua-checksums.json including global configuration",
			},
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "If a package's checksum configuration is disabled, download the asset and calculate the checksum",
			},
		},
		Description: `Create or Update .aqua-checksums.json.

e.g.
$ aqua update-checksum

By default aqua doesn't update .aqua-checksums.json of the global configuration.
If you want to update them too,
please set "-a" option.

$ aqua update-checksum -a
`,
		Action: runner.updateChecksumAction,
	}
}

func (runner *Runner) updateChecksumAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "update-checksum", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateChecksumCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	return ctrl.UpdateChecksum(c.Context, runner.LogE, param) //nolint:wrapcheck
}
