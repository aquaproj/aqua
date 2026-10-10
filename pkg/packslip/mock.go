package packslip

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

type MockVerifier struct {
	err error
}

func (m *MockVerifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, ps *registry.Packslip, art *template.Artifact, file *download.File, artifactPath, assetName string) error {
	return m.err
}

type MockExecutor struct {
	Err   error
	Param *ParamVerify
}

func (m *MockExecutor) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify) error {
	m.Param = param
	return m.Err
}
