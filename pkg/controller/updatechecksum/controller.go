package updatechecksum

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir            string
	configFinder       ConfigFinder
	configReader       ConfigReader
	registryInstaller  RegistryInstaller
	registryDownloader GitHubContentFileDownloader
	fs                 afero.Fs
	runtime            *runtime.Runtime
	chkDL              download.ChecksumDownloader
	downloader         download.ClientAPI
	prune              bool
}

func New(param *config.Param, configFinder ConfigFinder, configReader ConfigReader, registInstaller RegistryInstaller, fs afero.Fs, rt *runtime.Runtime, chkDL download.ChecksumDownloader, pkgDownloader download.ClientAPI, registDownloader GitHubContentFileDownloader) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		configFinder:       configFinder,
		configReader:       configReader,
		registryInstaller:  registInstaller,
		registryDownloader: registDownloader,
		fs:                 fs,
		runtime:            rt,
		chkDL:              chkDL,
		downloader:         pkgDownloader,
		prune:              param.Prune,
	}
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

type GitHubContentFileDownloader interface {
	DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error)
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
