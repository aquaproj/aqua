package completion

import (
	"github.com/aquaproj/aqua/v2/pkg/cli/util"
	"github.com/urfave/cli/v3"
)

type command struct {
	r   *util.Param
	cmd *cli.Command
}

func New(cmd *cli.Command) func(r *util.Param) *cli.Command {
	return func(r *util.Param) *cli.Command {
		i := &command{
			r:   r,
			cmd: cmd,
		}
		return &cli.Command{
			Name:  "completion",
			Usage: "Output shell completion script for bash, zsh, or fish",
			Description: `Output shell completion script for bash, zsh, or fish.
Source the output to enable completion.

e.g.

# .bashrc

if command -v aqua &> /dev/null; then
	eval "$(aqua completion bash)"
fi

# .zshrc

if command -v aqua &> /dev/null; then
	source <(aqua completion zsh)
fi

# fish

aqua completion fish > ~/.config/fish/completions/aqua.fish
`,
			Commands: []*cli.Command{
				{
					Name:   "bash",
					Usage:  "Output shell completion script for bash",
					Action: i.bash,
				},
				{
					Name:   "zsh",
					Usage:  "Output shell completion script for zsh",
					Action: i.zsh,
				},
				{
					Name:   "fish",
					Usage:  "Output shell completion script for fish",
					Action: i.fish,
				},
			},
		}
	}
}
