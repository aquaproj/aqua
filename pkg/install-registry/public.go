package registry

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/log"
)

type Installer interface {
	InstallRegistries(ctx context.Context, cfg *config.Config, cfgFilePath string) (map[string]*config.RegistryContent, error)
}

func New(rootDir config.RootDir, logger *log.Logger, downloader download.RegistryDownloader) Installer {
	return &installer{
		rootDir:            string(rootDir),
		registryDownloader: downloader,
		logger:             logger,
	}
}
