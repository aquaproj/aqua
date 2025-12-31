package cli

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/cli/cliargs"
	"github.com/aquaproj/aqua/v2/pkg/cli/cp"
	"github.com/aquaproj/aqua/v2/pkg/cli/exec"
	"github.com/aquaproj/aqua/v2/pkg/cli/generate"
	"github.com/aquaproj/aqua/v2/pkg/cli/genr"
	"github.com/aquaproj/aqua/v2/pkg/cli/info"
	"github.com/aquaproj/aqua/v2/pkg/cli/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/cli/install"
	"github.com/aquaproj/aqua/v2/pkg/cli/list"
	cpolicy "github.com/aquaproj/aqua/v2/pkg/cli/policy"
	"github.com/aquaproj/aqua/v2/pkg/cli/remove"
	"github.com/aquaproj/aqua/v2/pkg/cli/root"
	"github.com/aquaproj/aqua/v2/pkg/cli/token"
	"github.com/aquaproj/aqua/v2/pkg/cli/upc"
	"github.com/aquaproj/aqua/v2/pkg/cli/update"
	"github.com/aquaproj/aqua/v2/pkg/cli/updateaqua"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/cli/vacuum"
	"github.com/aquaproj/aqua/v2/pkg/cli/which"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

// Runner is the main CLI runner for aqua commands.
// It provides the entry point for executing aqua CLI operations.
type Runner struct{}

// newC is a function type that creates a new CLI command given a parameter and global args.
// It's used to construct commands in a consistent way throughout the CLI package.
type newC func(r *util.Param, globalArgs *cliargs.GlobalArgs) *cli.Command

// commands creates a slice of CLI commands by applying the given command constructors
// to the provided parameter. It takes a parameter and a variadic list of command
// constructor functions, returning the constructed commands.
func commands(param *util.Param, globalArgs *cliargs.GlobalArgs, newCs ...newC) []*cli.Command {
	cs := make([]*cli.Command, len(newCs))
	for i, newC := range newCs {
		cs[i] = newC(param, globalArgs)
	}
	return cs
}

func Run(ctx context.Context, logger *slogutil.Logger, env *urfave.Env) error {
	param := &util.Param{
		Stdin:   env.Stdin,
		Stdout:  env.Stdout,
		Stderr:  env.Stderr,
		Logger:  logger,
		Runtime: runtime.New(),
		Version: env.Version,
	}
	globalArgs := &cliargs.GlobalArgs{}
	return urfave.Command(env, &cli.Command{ //nolint:wrapcheck
		Name:           "aqua",
		Usage:          "Version Manager of CLI. https://aquaproj.github.io/",
		ExitErrHandler: exitErrHandlerFunc,
		Flags:          cliargs.GlobalFlags(globalArgs),
		Commands: commands(
			param,
			globalArgs,
			initcmd.New,
			install.New,
			generate.New,
			updateaqua.New,
			upc.New,
			update.New,
			which.New,
			info.New,
			remove.New,
			vacuum.New,
			token.New,
			cp.New,
			cpolicy.New,
			cpolicy.NewInitPolicy,
			exec.New,
			list.New,
			genr.New,
			root.New,
		),
	}).Run(ctx, env.Args)
}

// exitErrHandlerFunc handles exit errors for CLI commands.
// It provides special handling for the "exec" command by skipping the default
// error handling, allowing the exec command to manage its own exit codes.
func exitErrHandlerFunc(_ context.Context, cmd *cli.Command, err error) {
	if cmd.Name != "exec" {
		cli.HandleExitCoder(err)
		return
	}
}
