package checksumgetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type MockChecksumFileVerifier struct {
	Err error
}

func (v *MockChecksumFileVerifier) VerifyChecksumFileContent(_ context.Context, _ *slog.Logger, _ *config.Package, _ string, _ []byte) error {
	return v.Err
}

func (v *MockChecksumFileVerifier) VerifyAsset(_ context.Context, _ *slog.Logger, _ *config.Package, _, _ string, _ *runtime.Runtime) error {
	return v.Err
}
