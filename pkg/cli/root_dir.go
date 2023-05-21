package cli

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

func (runner *Runner) newRootDirCommand() *cli.Command {
	return &cli.Command{
		Name:  "root-dir",
		Usage: "Output the aqua root directory (AQUA_ROOT_DIR)",
		Description: `Output the aqua root directory (AQUA_ROOT_DIR)
e.g.

$ aqua root-dir
/home/foo/.local/share/aquaproj-aqua

$ export "PATH=$(aqua root-dir)/bin:PATH"
`,
		Action: runner.rootDirAction,
	}
}

func (runner *Runner) rootDirAction(c *cli.Context) error {
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

	fmt.Fprintln(runner.Stdout, config.GetRootDir(osenv.New()))
	return nil
}
