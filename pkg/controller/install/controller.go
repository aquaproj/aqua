package install

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/spf13/afero"
)

type Controller struct {
	packageInstaller   installpackage.Installer
	rootDir            string
	configFinder       ConfigFinder
	configReader       reader.ConfigReader
	registryInstaller  registry.Installer
	fs                 afero.Fs
	runtime            *runtime.Runtime
	tags               map[string]struct{}
	excludedTags       map[string]struct{}
	policyConfigFinder policy.ConfigFinder
	policyConfigReader policy.Reader
	skipLink           bool
	requireChecksum    bool
}

func New(param *config.Param, configFinder ConfigFinder, configReader reader.ConfigReader, registInstaller registry.Installer, pkgInstaller installpackage.Installer, fs afero.Fs, rt *runtime.Runtime, policyConfigReader policy.Reader, policyConfigFinder policy.ConfigFinder) *Controller {
	return &Controller{
		rootDir:            param.RootDir,
		configFinder:       configFinder,
		configReader:       configReader,
		registryInstaller:  registInstaller,
		packageInstaller:   pkgInstaller,
		fs:                 fs,
		runtime:            rt,
		skipLink:           param.SkipLink,
		tags:               param.Tags,
		excludedTags:       param.ExcludedTags,
		policyConfigReader: policyConfigReader,
		policyConfigFinder: policyConfigFinder,
		requireChecksum:    param.RequireChecksum,
	}
}
