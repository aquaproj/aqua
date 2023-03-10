package slsa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aquaproj/aqua/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier/verify"
)

type ExecutorImpl struct{}

func NewExecutor() *ExecutorImpl {
	return &ExecutorImpl{}
}

type Executor interface {
	Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error
}

type MockExecutor struct {
	Err error
}

func (mock *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	return mock.Err
}

func (exe *ExecutorImpl) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, provenancePath string) error {
	v := verify.VerifyArtifactCommand{
		ProvenancePath: provenancePath,
		SourceURI:      param.SourceURI,
		SourceTag:      &param.SourceTag,
	}
	for i := 0; i < 5; i++ {
		_, err := v.Exec(ctx, []string{param.ArtifactPath})
		if err == nil {
			return nil
		}
		if e := ctx.Err(); e != nil {
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		if i == 4 { //nolint:gomnd
			return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
		}
		logE.WithField("retry_count", i+1).Info("slsa-verifier failed. Retrying")
		if err := util.Wait(ctx, 1*time.Second); err != nil {
			return err //nolint:wrapcheck
		}
	}
	return nil
}
