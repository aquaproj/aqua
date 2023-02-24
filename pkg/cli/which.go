package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newWhichCommand() *cli.Command {
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
`,
		Action: runner.whichAction,
	}
}

func (runner *Runner) whichAction(c *cli.Context) error {
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
	if err := runner.setParam(c, "which", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl := controller.InitializeWhichCommandController(c.Context, param, http.DefaultClient, runner.Runtime)
	exeName, _, err := parseExecArgs(c.Args().Slice())
	if err != nil {
		return err
	}
	which, err := ctrl.Which(c.Context, runner.LogE, param, exeName)
	if err != nil {
		return err //nolint:wrapcheck
	}
	fmt.Fprintln(os.Stdout, which.ExePath)
	return nil
}
