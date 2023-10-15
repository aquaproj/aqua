package update

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	rgst "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	gh                RepositoriesService
	rootDir           string
	configFinder      ConfigFinder
	configReader      ConfigReader
	registryInstaller rgst.Installer
	fs                afero.Fs
	runtime           *runtime.Runtime
	requireChecksum   bool
	fuzzyGetter       FuzzyGetter
	fuzzyFinder       FuzzyFinder
	which             which.Controller
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
	ReadToUpdate(configFilePath string, cfg *aqua.Config) (map[string]*aqua.Config, error)
}

type FuzzyGetter interface {
	Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, useFinder bool) string
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	ListTags(ctx context.Context, owner string, repo string, opts *github.ListOptions) ([]*github.RepositoryTag, *github.Response, error)
}

func New(param *config.Param, gh RepositoriesService, configFinder ConfigFinder, configReader ConfigReader, registInstaller rgst.Installer, fs afero.Fs, rt *runtime.Runtime, fuzzyGetter FuzzyGetter, fuzzyFinder FuzzyFinder, whichController which.Controller) *Controller {
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
		fuzzyFinder:       fuzzyFinder,
		which:             whichController,
	}
}

type ConfigFinder interface {
	Find(wd, configFilePath string, globalConfigFilePaths ...string) (string, error)
}
