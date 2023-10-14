package update

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	gh                RepositoriesService
	rootDir           string
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller registry.Installer
	fs                afero.Fs
	runtime           *runtime.Runtime
	requireChecksum   bool
	fuzzyGetter       FuzzyGetter
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
	ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error)
}

type FuzzyGetter interface {
	Get(ctx context.Context, logE *logrus.Entry, pkg *fuzzyfinder.Package, useFinder bool) string
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

func New(param *config.Param, gh RepositoriesService, configFinder ConfigFinder, configReader ConfigReader, registInstaller registry.Installer, fs afero.Fs, rt *runtime.Runtime, fuzzyGetter FuzzyGetter) *Controller {
	return &Controller{
		gh:                gh,
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		fs:                fs,
		runtime:           rt,
		requireChecksum:   param.RequireChecksum,
		fuzzyGetter:       fuzzyGetter,
	}
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}
