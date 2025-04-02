package upc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name: "update-checksum",
		Aliases: []string{
			"upc",
		},
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
		Action: i.action,
	}
}

func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(cmd, i.r.LogE, "update-checksum", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeUpdateChecksumCommandController(ctx, param, http.DefaultClient, i.r.Runtime)
	return ctrl.UpdateChecksum(ctx, i.r.LogE, param) //nolint:wrapcheck
}
