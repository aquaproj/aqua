package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newSetShellCommand() *cli.Command {
	return &cli.Command{
		Name:  "set-shell",
		Usage: "Output a script to set the shell",
		Description: `Output a script to set the shell

Please run this command in the shell script such as .bashrc and .zshrc.

eval "$(aqua set-shell <shell name>)"

The supported shell names are:

- zsh
`,
		Action: r.setShellAction,
		Flags:  []cli.Flag{},
	}
}

func (r *Runner) setShellAction(c *cli.Context) error {
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

	param := &config.Param{
		Ppid:              os.Getppid(),
		EnvPath:           os.Getenv("PATH"),
		PathListSeparator: string(os.PathListSeparator),
	}
	if err := r.setParam(c, "set-shell", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeSetShellCommandController(c.Context, param, http.DefaultClient, r.Runtime, r.Stdout)
	if err != nil {
		return fmt.Errorf("initialize a SetShellController: %w", err)
	}
	return ctrl.SetShell(c.Context, r.LogE, param, c.Args().First()) //nolint:wrapcheck
}
