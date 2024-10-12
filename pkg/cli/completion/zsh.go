package completion

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func (cm *command) zsh(*cli.Context) error {
	// https://github.com/urfave/cli/blob/main/autocomplete/zsh_autocomplete
	// https://github.com/urfave/cli/blob/947f9894eef4725a1c15ed75459907b52dde7616/autocomplete/zsh_autocomplete
	fmt.Fprintln(cm.r.Stdout, `#compdef aqua

_aqua() {
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

if [ "$funcstack[1]" = "_aqua" ]; then
  _aqua "$@"
else
  compdef _aqua aqua
fi`)
	return nil
}
