package exec

import "context"

type Executor interface {
	Exec(ctx context.Context, exePath string, args ...string) (int, error)
	ExecXSys(exePath string, args ...string) error
}
