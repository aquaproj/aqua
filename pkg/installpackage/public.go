package installpackage

import (
	"github.com/aquaproj/aqua/v2/pkg/exec"
)

type Executor interface {
	Exec(cmd *exec.Cmd, param *exec.ParamRun) (int, error)
	ExecAndOutputWhenFailure(cmd *exec.Cmd) (int, error)
}
