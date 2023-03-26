package cli

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newUpdateChecksumCommand() *cli.Command {
	return &cli.Command{
		Name:  "update-checksum",
		Usage: "Create or Update aqua-checksums.json",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Create or Update all aqua-checksums.json including global configuration",
			},
			&cli.BoolFlag{
				Name:  "deep",
				Usage: "This flag was deprecated and had no meaning from aqua v2.0.0. This flag will be removed in aqua v3.0.0. https://github.com/aquaproj/aqua/issues/1769",
			},
			&cli.BoolFlag{
				Name:  "prune",
				Usage: "Remove unused checksums",
			},
		},
		Description: `Create or Update aqua-checksums.json.

e.g.
$ aqua update-checksum

By default aqua doesn't update aqua-checksums.json of the global configuration.
If you want to update them too,
please set "-a" option.

$ aqua update-checksum -a

By default, aqua update-checksum doesn't remove existing checksums even if they aren't unused.
If -prune option is set, aqua unused checksums would be removed.

$ aqua update-checksum -prune
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
