package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, cfg *aqua.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*registry.Config, error)
}

type MockRegistryInstaller struct {
	M   map[string]*registry.Config
	Err error
}

func (inst *MockRegistryInstaller) InstallRegistries(ctx context.Context, cfg *aqua.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*registry.Config, error) {
	return inst.M, inst.Err
}
