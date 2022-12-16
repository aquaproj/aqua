package registry

import (
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/spf13/afero"
)

func New(param *config.Param, downloader domain.GitHubContentFileDownloader, fs afero.Fs, rt *runtime.Runtime, cos CosignVerifier) *Installer {
	return &Installer{
		param:              param,
		registryDownloader: downloader,
		fs:                 fs,
		rt:                 rt,
		cosign:             cos,
	}
}
