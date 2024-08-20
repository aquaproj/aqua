package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller"
	"github.com/urfave/cli/v2"
)

func (r *Runner) newOutputShellCommand() *cli.Command {
	return &cli.Command{
		Name:  "output-shell",
		Usage: "Output a script to set the shell",
		Description: `Output a script to set the shell

aqua set-shell command executes this command.
`,
		Action: r.outputShellAction,
		Flags:  []cli.Flag{},
	}
}

func (r *Runner) outputShellAction(c *cli.Context) error {
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
	if err := r.setParam(c, "output-shell", param); err != nil {
		return fmt.Errorf("parse the command line arguments: %w", err)
	}
	ctrl, err := controller.InitializeOutputShellCommandController(c.Context, param, http.DefaultClient, r.Runtime, r.Stdout)
	if err != nil {
		return fmt.Errorf("initialize a OutputShellController: %w", err)
	}
	return ctrl.OutputShell(c.Context, r.LogE, param) //nolint:wrapcheck
}
