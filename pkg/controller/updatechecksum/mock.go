package updatechecksum

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

type MockConfigFinder struct {
	Files []string
}

func (f *MockConfigFinder) Finds(wd, configFilePath string) []string {
	return f.Files
}

type MockChecksumFileVerifier struct {
	Err error
}

func (v *MockChecksumFileVerifier) VerifyChecksumFileContent(_ context.Context, _ *slog.Logger, _ *config.Package, _ string, _ []byte) error {
	return v.Err
}
