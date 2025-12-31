// Package root implements the aqua root-dir command for displaying the aqua root directory.
// The root-dir command outputs the path to the aqua root directory (AQUA_ROOT_DIR)
// which is used for storing installed packages and configuration files.
package root

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the root-dir command.
type Args struct {
	*cliargs.GlobalArgs
}

// command holds the parameters and configuration for the root-dir command.
type command struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for displaying the root directory.
// The returned command outputs the aqua root directory path which can be
// used for PATH configuration and understanding aqua's file structure.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	i := &command{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:  "root-dir",
		Usage: "Output the aqua root directory (AQUA_ROOT_DIR)",
		Description: `Output the aqua root directory (AQUA_ROOT_DIR)
e.g.

$ aqua root-dir
/home/foo/.local/share/aquaproj-aqua

$ export "PATH=$(aqua root-dir)/bin:PATH"
`,
		Action: i.action,
	}
}

// action implements the main logic for the root-dir command.
// It outputs the aqua root directory path to stdout for use in
// shell scripts and PATH configuration.
func (i *command) action(_ context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(i.args.Trace, i.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	fmt.Fprintln(i.r.Stdout, config.GetRootDir(osenv.New()))
	return nil
}
