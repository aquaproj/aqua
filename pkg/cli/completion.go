package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func (r *Runner) newCompletionCommand() *cli.Command {
	// https://github.com/aquaproj/aqua/pull/859
	// https://cli.urfave.org/v2/#bash-completion
	return &cli.Command{
		Name:  "completion",
		Usage: "Output shell completion script for bash or zsh",
		Description: `Output shell completion script for bash or zsh
Run these commands in .bash_profile or .zprofile
e.g.
.bash_profile

if command -v aqua &> /dev/null; then source <(aqua completion bash); fi

.zprofile

if command -v aqua &> /dev/null; then source <(aqua completion zsh); fi
`,
		Subcommands: []*cli.Command{
			{
				Name:   "bash",
				Usage:  "Output shell completion script for bash",
				Action: r.bashCompletionAction,
			},
			{
				Name:   "zsh",
				Usage:  "Output shell completion script for zsh",
				Action: r.zshCompletionAction,
			},
		},
	}
}

func (r *Runner) bashCompletionAction(*cli.Context) error {
	// https://github.com/urfave/cli/blob/main/autocomplete/bash_autocomplete
	// https://github.com/urfave/cli/blob/c3f51bed6fffdf84227c5b59bd3f2e90683314df/autocomplete/bash_autocomplete#L5-L20
	fmt.Fprintln(r.Stdout, `
_cli_bash_autocomplete() {
  if [[ "${COMP_WORDS[0]}" != "source" ]]; then
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "$cur" == "-"* ]]; then
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} ${cur} --generate-bash-completion )
    else
      opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    fi
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
  fi
}

complete -o bashdefault -o default -o nospace -F _cli_bash_autocomplete aqua`)
	return nil
}

func (r *Runner) zshCompletionAction(*cli.Context) error {
	// https://github.com/urfave/cli/blob/main/autocomplete/zsh_autocomplete
	// https://github.com/urfave/cli/blob/947f9894eef4725a1c15ed75459907b52dde7616/autocomplete/zsh_autocomplete
	fmt.Fprintln(r.Stdout, `
#compdef aqua

_cli_zsh_autocomplete() {
  local -a opts
  local cur
  cur=${words[-1]}
  if [[ "$cur" == "-"* ]]; then
    opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
  else
    opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
  fi

  if [[ "${opts[1]}" != "" ]]; then
    _describe 'values' opts
  else
    _files
  fi
}

compdef _cli_zsh_autocomplete aqua`)
	return nil
}
