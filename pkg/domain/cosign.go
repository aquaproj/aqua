package domain

import (
	"context"

	"github.com/sirupsen/logrus"
)

// import (
// 	"context"
//
// 	"github.com/aquaproj/aqua/pkg/cosign"
// )
//
// type CosignVerifier interface {
// 	Verify(ctx context.Context, param *cosign.ParamVerify) error
// 	HasCosign() bool
// }
//
// type MockCosignVerifier struct {
// 	hasCosign bool
// 	err       error
// }
//
// func (cos *MockCosignVerifier) HasCosign() bool {
// 	return cos.hasCosign
// }
//
// func (cos *MockCosignVerifier) Verify(ctx context.Context, param *cosign.ParamVerify) error {
// 	return cos.err
// }

type CosignInstaller interface {
	InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error
}

type MockCosignInstaller struct {
	err error
}

func (mock *MockCosignInstaller) InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error {
	return mock.err
}
