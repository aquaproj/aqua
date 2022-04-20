package registry

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/sirupsen/logrus"
)

type Installer interface {
	InstallRegistries(ctx context.Context, cfg *config.Config, cfgFilePath string, logE *logrus.Entry) (map[string]*config.RegistryContent, error)
}

func New(rootDir config.RootDir, downloader download.RegistryDownloader) Installer {
	return &installer{
		rootDir:            string(rootDir),
		registryDownloader: downloader,
	}
}
