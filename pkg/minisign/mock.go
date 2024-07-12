package minisign

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/sirupsen/logrus"
)

type MockVerifier struct {
	err error
}

func (m *MockVerifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, ms *registry.Minisign, art *template.Artifact, file *download.File, param *ParamVerify) error {
	return m.err
}

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logE *logrus.Entry, param *ParamVerify, signature string) error {
	return m.Err
}
