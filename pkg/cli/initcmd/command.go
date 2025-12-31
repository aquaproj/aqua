// Package initcmd implements the aqua init command for creating configuration files.
// The init command creates new aqua configuration files with optional
// directory structure and import configurations for project initialization.
package initcmd

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/urfave/cli/v3"
)

// Args holds command-line arguments for the init command.
type Args struct {
	*cliargs.GlobalArgs

	UseImportDir bool
	ImportDir    string
	CreateDir    bool
	FilePath     []string
}

// initCommand holds the parameters and configuration for the init command.
type initCommand struct {
	r    *util.Param
	args *Args
}

// New creates and returns a new CLI command for project initialization.
// The returned command creates aqua configuration files with options
// for directory structure and import configurations.
func New(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command {
	args := &Args{
		GlobalArgs: globalArgs,
	}
	ic := &initCommand{
		r:    r,
		args: args,
	}
	return &cli.Command{
		Name:      "init",
		Usage:     "Create a configuration file if it doesn't exist",
		ArgsUsage: `[<created file path. The default value is "aqua.yaml">]`,
		Description: `Create a configuration file if it doesn't exist
e.g.
$ aqua init # create "aqua.yaml"
$ aqua init foo.yaml # create foo.yaml
$ aqua init -u # Replace "packages:" with "import_dir: imports"
$ aqua init -i <directory path> # Replace "packages:" with "import_dir: <directory path>"
$ aqua init -d # Create a directory "aqua" and create "aqua/aqua.yaml"
`,
		Action: ic.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "use-import-dir",
				Aliases:     []string{"u"},
				Usage:       "Use import_dir",
				Destination: &args.UseImportDir,
			},
			&cli.StringFlag{
				Name:        "import-dir",
				Aliases:     []string{"i"},
				Usage:       "import_dir",
				Destination: &args.ImportDir,
			},
			&cli.BoolFlag{
				Name:        "create-dir",
				Aliases:     []string{"d"},
				Usage:       "Create a directory named aqua and create aqua.yaml in it",
				Destination: &args.CreateDir,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "file_path",
				Min:         0,
				Max:         1,
				Destination: &args.FilePath,
			},
		},
	}
}

// action implements the main logic for the init command.
// It creates configuration files and directory structures based on
// the provided command line options and arguments.
func (ic *initCommand) action(ctx context.Context, _ *cli.Command) error {
	profiler, err := profile.Start(ic.args.Trace, ic.args.CPUProfile)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	param := &config.Param{}
	if err := util.SetParam(ic.args.GlobalArgs, ic.r.Logger, param, ic.r.Version); err != nil {
		return fmt.Errorf("set param: %w", err)
	}
	ctrl := controller.InitializeInitCommandController(ctx, ic.r.Logger.Logger, param)
	cParam := &initcmd.Param{
		IsDir:     ic.args.CreateDir,
		ImportDir: ic.args.ImportDir,
	}
	if cParam.ImportDir == "" && ic.args.UseImportDir {
		cParam.ImportDir = "imports"
	}
	filePath := ""
	if len(ic.args.FilePath) > 0 {
		filePath = ic.args.FilePath[0]
	}
	return ctrl.Init(ctx, ic.r.Logger.Logger, filePath, cParam) //nolint:wrapcheck
}
