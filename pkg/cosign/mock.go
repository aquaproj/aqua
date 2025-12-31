package cosign

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

func (v *MockVerifier) Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error {
	return v.err
}
