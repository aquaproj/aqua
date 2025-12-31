// Package cp implements the aqua cp command for copying executable files.
// The cp command copies executable files from aqua's installation directory
// to a specified destination directory, allowing for easy distribution
// and deployment of installed tools.
package cp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the cp command.
type Args struct {
	*cliargs.GlobalArgs

	Output      string
	All         bool
	Tags        string
	ExcludeTags string
	Commands    []string
}

// command holds the parameters and configuration for the cp command.
type command struct {
	r *util.Param
}

// New creates and returns a new CLI command for copying executable files.
// The returned command allows users to copy installed tool executables
// to a specified directory with various filtering options.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r: r,
	}
	return &cli.Command{
		Name:      "cp",
		Usage:     "Copy executable files in a directory",
		ArgsUsage: `<command name> [<command name> ...]`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "o",
				Value:       "dist",
				Usage:       "destination directory",
				Destination: &args.Output,
			},
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "install all aqua configuration packages",
				Destination: &args.All,
			},
			&cli.StringFlag{
				Name:        "tags",
				Aliases:     []string{"t"},
				Usage:       "filter installed packages with tags",
				Destination: &args.Tags,
			},
			&cli.StringFlag{
				Name:        "exclude-tags",
				Usage:       "exclude installed packages with tags",
				Destination: &args.ExcludeTags,
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
		Action: func(ctx context.Context, _ *cli.Command) error {
			return i.action(ctx, args)
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "commands",
				Min:         0,
				Max:         -1,
				Destination: &args.Commands,
			},
		},
	}
}

// action implements the main logic for the cp command.
// It initializes the copy controller and executes the file copying operation
// based on the provided command line arguments and configuration.
func (i *command) action(ctx context.Context, args *Args) error {
	profiler, err := profile.Start(args.Trace, args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	param.SkipLink = true
	param.Dest = args.Output
	param.All = args.All
	param.Tags = parseTags(args.Tags)
	param.ExcludedTags = parseTags(args.ExcludeTags)
	param.Commands = args.Commands
	ctrl, err := controller.InitializeCopyCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	if err != nil {
		return fmt.Errorf("initialize a CopyController: %w", err)
	}
	if err := ctrl.Copy(ctx, i.r.Logger.Logger, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func parseTags(tags string) map[string]struct{} {
	tagsM := map[string]struct{}{}
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		tagsM[tag] = struct{}{}
	}
	return tagsM
}
