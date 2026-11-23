package registry

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

type MockInstaller struct {
	M   map[string]*registry.Config
	Err error
}

func (m *MockInstaller) InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error) {
	return m.M, m.Err
}

func (m *MockInstaller) InstallRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, cfgFilePath string, checksums *checksum.Checksums) (*registry.Config, error) {
	return m.M[regist.Name], m.Err
}

func (m *MockInstaller) InstallLockedRegistry(ctx context.Context, logger *slog.Logger, regist *aqua.Registry, commit string, chk *checksum.Checksum) (*registry.Config, error) {
	return m.M[regist.Name], m.Err
}
