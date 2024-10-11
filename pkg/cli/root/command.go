package root

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cpuprofile"
	"github.com/aquaproj/aqua/v2/pkg/cli/tracer"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v2"
)

type command struct {
	r *util.Param
}

func New(r *util.Param) *cli.Command {
	i := &command{
		r: r,
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

func (i *command) action(c *cli.Context) error {
	tracer, err := tracer.Start(c.String("trace"))
	if err != nil {
		return err
	}
	defer tracer.Stop()

	cpuProfiler, err := cpuprofile.Start(c.String("cpu-profile"))
	if err != nil {
		return err
	}
	defer cpuProfiler.Stop()

	fmt.Fprintln(i.r.Stdout, config.GetRootDir(osenv.New()))
	return nil
}
