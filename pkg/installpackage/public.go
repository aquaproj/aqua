package installpackage

import (
	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type Executor interface {
	Exec(cmd *osexec.Cmd) (int, error)
	ExecStderr(cmd *osexec.Cmd) (int, error)
	ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error)
}
