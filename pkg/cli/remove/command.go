package remove

import (
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
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

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
	}
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
			&cli.StringFlag{
				Name:    "mode",
				Aliases: []string{"m"},
				EnvVars: []string{"AQUA_REMOVE_MODE"},
				Usage:   "Removed target modes. l: link, p: package",
			},
			&cli.BoolFlag{
				Name:  "i",
				Usage: "Select packages with a Fuzzy Finder",
			},
		},
		Description: description,
		Action:      i.action,
	}
}

func (i *command) action(c *cli.Context) error {
	profiler, err := profile.Start(c)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	mode, err := parseRemoveMode(c.String("mode"))
	if err != nil {
		return fmt.Errorf("parse the mode option: %w", err)
	}

	param := &config.Param{}
	if err := util.SetParam(c, i.r.LogE, "remove", param, i.r.LDFlags); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	param.SkipLink = true
	ctrl := controller.InitializeRemoveCommandController(c.Context, i.r.LogE, param, http.DefaultClient, i.r.Runtime, mode)
	if err := ctrl.Remove(c.Context, i.r.LogE, param); err != nil {
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
