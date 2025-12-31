// Package which implements the aqua which command for locating executable files.
// The which command outputs the absolute file path of installed tools,
// helping users understand which version and location of a tool is being used.
package which

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the which command.
type Args struct {
	*cliargs.GlobalArgs

	ShowVersion bool
	WhichArgs   []string
}

// command holds the parameters and configuration for the which command.
type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for locating executables.
// The returned command provides functionality to find the absolute path
// of installed tools managed by aqua.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:      "which",
		Usage:     "Output the absolute file path of the given command",
		ArgsUsage: `<command name>`,
		Description: `Output the absolute file path of the given command
e.g.
$ aqua which gh
/home/foo/.aqua/pkgs/github_release/github.com/cli/cli/v2.4.0/gh_2.4.0_macOS_amd64.tar.gz/gh_2.4.0_macOS_amd64/bin/gh

If the command isn't found in the configuration files, aqua searches the command in the environment variable PATH

$ aqua which ls
/bin/ls

If the command isn't found, exits with non zero exit code.

$ aqua which foo
FATA[0000] aqua failed                                   aqua_version=0.8.6 error="command is not found" exe_name=foo program=aqua

If you want the package version, "--version" option is useful.

$ aqua which --version gh
v2.4.0
`,
		Action: i.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "version",
				Aliases:     []string{"v"},
				Usage:       "Output the given package version",
				Destination: &args.ShowVersion,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "which_args",
				Min:         0,
				Max:         -1,
				Destination: &args.WhichArgs,
			},
		},
	}
}

func (i *command) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(i.args.GlobalArgs, i.r.Logger, param, i.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	param.ShowVersion = i.args.ShowVersion
	ctrl := controller.InitializeWhichCommandController(ctx, i.r.Logger.Logger, param, http.DefaultClient, i.r.Runtime)
	exeName, _, err := ParseExecArgs(i.args.WhichArgs)
	if err != nil {
		return err
	}
	logger := i.r.Logger.With("exe_name", exeName)
	which, err := ctrl.Which(ctx, logger, param, exeName)
	if err != nil {
		return slogerr.With(err, "exe_name", exeName) //nolint:wrapcheck
	}
	if !param.ShowVersion {
		fmt.Fprintln(os.Stdout, which.ExePath)
		return nil
	}
	if which.Package == nil {
		return slogerr.With(errors.New("aqua can't get the command version because the command isn't managed by aqua"), "exe_name", exeName) //nolint:wrapcheck
	}
	fmt.Fprintln(os.Stdout, which.Package.Package.Version)
	return nil
}

var errCommandIsRequired = errors.New("command is required")

func ParseExecArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", nil, errCommandIsRequired
	}
	return filepath.Base(args[0]), args[1:], nil
}
