package exec

import "context"

type Executor interface {
	ExecXSysWithEnvs(exePath string, args, envs []string) error
	ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error)
}
