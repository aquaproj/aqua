package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func (runner *Runner) newCompletionCommand() *cli.Command {
	return &cli.Command{
		Name:  "completion",
		Usage: "Output shell completion script for bash or zsh",
		Subcommands: []*cli.Command{
			{
				Name:   "bash",
				Usage:  "Output shell completion script for bash",
				Action: runner.bashCompletionAction,
			},
			{
				Name:   "zsh",
				Usage:  "Output shell completion script for zsh",
				Action: runner.zshCompletionAction,
			},
		},
	}
}

func (runner *Runner) bashCompletionAction(c *cli.Context) error {
	fmt.Println(`
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

func (runner *Runner) zshCompletionAction(c *cli.Context) error {
	fmt.Println(`
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
