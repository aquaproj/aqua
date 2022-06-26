package registry

import (
	"context"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/download"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Installer interface {
	InstallRegistries(ctx context.Context, cfg *aqua.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*registry.Config, error)
}

func New(param *config.Param, downloader download.RegistryDownloader, fs afero.Fs) Installer {
	return &installer{
		param:              param,
		registryDownloader: downloader,
		fs:                 fs,
	}
}
