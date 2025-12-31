package update

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	gh                RepositoriesService
	rootDir           string
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller RegistryInstaller
	fs                afero.Fs
	runtime           *runtime.Runtime
	fuzzyGetter       FuzzyGetter
	fuzzyFinder       FuzzyFinder
	which             WhichController
}

type WhichController interface {
	Which(ctx context.Context, logger *slog.Logger, param *config.Param, exeName string) (*which.FindResult, error)
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
	ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error)
}

type FuzzyGetter interface {
	Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) string
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

func New(param *config.Param, gh RepositoriesService, configFinder ConfigFinder, configReader ConfigReader, registryInstaller RegistryInstaller, fs afero.Fs, rt *runtime.Runtime, fuzzyGetter FuzzyGetter, fuzzyFinder FuzzyFinder, whichController WhichController) *Controller {
	return &Controller{
		gh:                gh,
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registryInstaller,
		fs:                fs,
		runtime:           rt,
		fuzzyGetter:       fuzzyGetter,
		fuzzyFinder:       fuzzyFinder,
		which:             whichController,
	}
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
