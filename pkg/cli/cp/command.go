package cp

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
	if err := util.SetParam(cmd, i.r.LogE, "cp", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	param.SkipLink = true
	ctrl, err := controller.InitializeCopyCommandController(ctx, i.r.LogE, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize a CopyController: %w", err)
	}
	if err := ctrl.Copy(ctx, i.r.LogE, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
