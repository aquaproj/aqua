package installpackage

import (
	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type Executor interface {
	ExecStderr(cmd *osexec.Cmd) (int, error)
	ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error)
}
