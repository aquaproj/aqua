package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type MockCargoPackageInstaller struct {
	Err error
}

func (m *MockCargoPackageInstaller) Install(ctx context.Context, logE *logrus.Entry, crate, version, root string, opts *registry.Cargo) error {
	return m.Err
}
