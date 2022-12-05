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
	Opts   []string
	Target string
}

func (verifier *Verifier) Verify(ctx context.Context, param *ParamVerify) error {
	_, err := verifier.executor.ExecWithEnvs(ctx, "cosign", append([]string{"verify-blob"}, append(param.Opts, param.Target)...), []string{"COSIGN_EXPERIMENTAL=1"})
	if err != nil {
		return fmt.Errorf("verify with cosign: %w", err)
	}
	return nil
}
