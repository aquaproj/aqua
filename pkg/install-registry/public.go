package registry

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Installer interface {
	InstallRegistries(ctx context.Context, cfg *aqua.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*registry.Config, error)
}

func New(param *config.Param, downloader domain.RegistryDownloader, fs afero.Fs) Installer {
	return &installer{
		param:              param,
		registryDownloader: downloader,
		fs:                 fs,
	}
}
