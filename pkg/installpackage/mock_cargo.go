package installpackage

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

type MockCargoPackageInstaller struct {
	Err error
}

func (m *MockCargoPackageInstaller) Install(ctx context.Context, logger *slog.Logger, crate, version, root string, opts *registry.Cargo) error {
	return m.Err
}
