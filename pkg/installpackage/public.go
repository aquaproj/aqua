package installpackage

import (
	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type Executor interface {
	Exec(cmd *osexec.Cmd, param *osexec.ParamRun) (int, error)
	ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error)
}
