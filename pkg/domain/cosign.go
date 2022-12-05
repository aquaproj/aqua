package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/cosign"
)

type CosignVerifier interface {
	Verify(ctx context.Context, param *cosign.ParamVerify) error
	HasCosign() bool
}

type MockCosignVerifier struct {
	hasCosign bool
	err       error
}

func (cos *MockCosignVerifier) HasCosign() bool {
	return cos.hasCosign
}

func (cos *MockCosignVerifier) Verify(ctx context.Context, param *cosign.ParamVerify) error {
	return cos.err
}
