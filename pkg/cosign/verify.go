package cosign

import (
	"context"
	"fmt"
)

type Verifier struct {
	executor Executor
}

type Executor interface {
	ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error)
}

func NewVerifier(executor Executor) *Verifier {
	return &Verifier{
		executor: executor,
	}
}

type ParamVerify struct {
	CosignExperimental bool
	Opts               []string
	Target             string
}

func (verifier *Verifier) Verify(ctx context.Context, param *ParamVerify) error {
	envs := []string{}
	if param.CosignExperimental {
		envs = []string{"COSIGN_EXPERIMENTAL=1"}
	}
	_, err := verifier.executor.ExecWithEnvs(ctx, "cosign", append([]string{"verify-blob"}, append(param.Opts, param.Target)...), envs)
	if err != nil {
		return fmt.Errorf("verify with cosign: %w", err)
	}
	return nil
}
