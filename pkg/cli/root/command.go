package root

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/profile"
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
	"github.com/urfave/cli/v3"
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

func (i *command) action(ctx context.Context, cmd *cli.Command) error {
	profiler, err := profile.Start(cmd)
	if err != nil {
		return fmt.Errorf("start CPU Profile or tracing: %w", err)
	}
	defer profiler.Stop()

	fmt.Fprintln(i.r.Stdout, config.GetRootDir(osenv.New()))
	return nil
}
