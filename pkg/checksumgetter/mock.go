package checksumgetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

type MockChecksumFileVerifier struct {
	Err error
}

func (v *MockChecksumFileVerifier) VerifyChecksumFileContent(_ context.Context, _ *slog.Logger, _ *config.Package, _ string, _ []byte) error {
	return v.Err
}
