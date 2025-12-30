package slsa

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

func (m *MockVerifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error {
	return m.err
}

type MockExecutor struct {
	Err error
}

func (m *MockExecutor) Verify(ctx context.Context, logger *slog.Logger, param *ParamVerify, provenancePath string) error {
	return m.Err
}
