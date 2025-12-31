// Package remove implements the aqua remove command for uninstalling packages.
// The remove command uninstalls packages from the aqua installation directory,
// providing options to remove package files and symbolic links.
package remove

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v3"
)

const description = `Uninstall packages.

e.g.
$ aqua rm --all
$ aqua rm cli/cli direnv/direnv tfcmt # Package names and command names

Note that this command remove files from AQUA_ROOT_DIR/pkgs, but doesn't remove packages from aqua.yaml and doesn't remove files from AQUA_ROOT_DIR/bin and AQUA_ROOT_DIR/bat.

If you want to uninstall packages of non standard registry, you need to specify the registry name too.

e.g.
$ aqua rm foo,suzuki-shunsuke/foo

By default, this command removes only packages from the pkgs directory and doesn't remove links from the bin directory.
You can change this behaviour by specifying the -mode flag.
The value of -mode is a string containing characters "l" and "p".
The order of the characters doesn't matter.

$ aqua rm -m l cli/cli # Remove only links
$ aqua rm -m pl cli/cli # Remove links and packages

Limitation:
"http" and "go_install" packages can't be removed.
`

// Args holds command-line arguments for the remove command.
type Args struct {
	*cliargs.GlobalArgs

	All         bool
	Mode        string
	Interactive bool
	Packages    []string
}

type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for removing packages.
// The returned command provides package uninstallation functionality
// with options for removing files and links based on specified modes.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Uninstall packages",
		ArgsUsage: `[<registry name>,]<package name> [...]`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "uninstall all packages",
				Destination: &args.All,
			},
			&cli.StringFlag{
				Name:        "mode",
				Aliases:     []string{"m"},
				Sources:     cli.EnvVars("AQUA_REMOVE_MODE"),
				Usage:       "Removed target modes. l: link, p: package",
				Destination: &args.Mode,
			},
			&cli.BoolFlag{
				Name:        "i",
				Usage:       "Select packages with a Fuzzy Finder",
				Destination: &args.Interactive,
			},
		},
		Description: description,
		Action:      i.action,
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "packages",
				Min:         0,
				Max:         -1,
				Destination: &args.Packages,
			},
		},
	}
}

// action implements the main logic for the remove command.
// It initializes the remove controller and executes package removal
// based on the provided command line arguments and mode settings.
func (i *command) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	mode, err := parseRemoveMode(i.args.Mode)
	if err != nil {
		return fmt.Errorf("parse the mode option: %w", err)
	}

	param := &config.Param{}
	if err := util.SetParam(i.args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	param.SkipLink = true
	param.All = i.args.All
	param.SelectVersion = i.args.Interactive
	param.Args = i.args.Packages
	ctrl := controller.InitializeRemoveCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime, mode)
	if err := ctrl.Remove(ctx, i.r.Logger.Logger, param); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

func parseRemoveMode(target string) (*config.RemoveMode, error) {
	if target == "" {
		return &config.RemoveMode{
			Package: true,
		}, nil
	}
	t := &config.RemoveMode{}
	for _, c := range target {
		switch c {
		case 'l':
			t.Link = true
		case 'p':
			t.Package = true
		default:
			return nil, fmt.Errorf("invalid mode: %c", c)
		}
	}
	return t, nil
}
