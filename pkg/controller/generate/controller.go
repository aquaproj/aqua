package generate

import (
	"io"
	"os"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	rgst "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/spf13/afero"
)

type Controller struct {
	stdin             io.Reader
	github            RepositoriesService
	registryInstaller rgst.Installer
	configFinder      ConfigFinder
	configReader      reader.ConfigReader
	fuzzyFinder       FuzzyFinder
	fs                afero.Fs
	outputter         Outputter
	cargoClient       cargo.Client
}

type VersionSelector interface {
	Find(versions []*fuzzyfinder.Version, hasPreview bool) (int, error)
}

type FuzzyFinder interface {
	Find(items []fuzzyfinder.Item, hasPreview bool) (int, error)
	FindMulti(items []fuzzyfinder.Item, hasPreview bool) ([]int, error)
}

func New(configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller rgst.Installer, gh RepositoriesService, fs afero.Fs, fuzzyFinder FuzzyFinder, cargoClient cargo.Client) *Controller {
	return &Controller{
		stdin:             os.Stdin,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		github:            gh,
		fs:                fs,
		fuzzyFinder:       fuzzyFinder,
		cargoClient:       cargoClient,
		outputter:         output.New(os.Stdout, fs),
	}
}
