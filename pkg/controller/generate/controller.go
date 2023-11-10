package generate

import (
	"context"
	"io"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	stdin             io.Reader
	github            RepositoriesService
	registryInstaller RegistryInstaller
	configFinder      ConfigFinder
	configReader      ConfigReader
	fuzzyFinder       FuzzyFinder
	fs                afero.Fs
	outputter         Outputter
	fuzzyGetter       FuzzyGetter
}

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type FuzzyGetter interface {
	Get(ctx context.Context, logE *logrus.Entry, pkg *registry.PackageInfo, currentVersion string, useFinder bool, limit int) string
}

type FuzzyFinder interface {
	Find(items []*fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []*fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func New(configFinder ConfigFinder, configReader ConfigReader, registInstaller RegistryInstaller, gh RepositoriesService, fs afero.Fs, fuzzyFinder FuzzyFinder, fuzzyGetter FuzzyGetter) *Controller {
	return &Controller{
		stdin:             os.Stdin,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		github:            gh,
		fs:                fs,
		fuzzyFinder:       fuzzyFinder,
		outputter:         output.New(os.Stdout, fs),
		fuzzyGetter:       fuzzyGetter,
	}
}

type RegistryInstaller interface {
	InstallRegistries(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, cfgFilePath string, checksums *checksum.Checksums) (map[string]*registry.Config, error)
}
