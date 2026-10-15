// Package migrate brings a configuration to the shape a newer aqua wants.
//
// A migration is a one-off. It changes what a configuration says rather than bringing it
// up to date with something that moved -- which is 'aqua fix' -- and it changes what is
// installed for nobody, which is 'aqua up'. Each one is its own command, because each
// carries its own decision: whether to keep aqua-checksums.json for whoever hasn't
// upgraded is not a question the others ask, and a flag for it belongs where it is asked.
//
// They come with an expiry nobody has to enforce: once everybody has run one, the code for
// it can go.
package migrate

import (
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

// ConfigFinder lists the aqua.yaml files that apply to a directory.
type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

// ConfigReader reads an aqua.yaml, resolving its imports.
type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

// Controller migrates configurations.
type Controller struct {
	configFinder ConfigFinder
	configReader ConfigReader
}

// New creates a Controller.
func New(configFinder ConfigFinder, configReader ConfigReader) *Controller {
	return &Controller{configFinder: configFinder, configReader: configReader}
}

// ChecksumArgs are the checksum migration's options.
//
// Each migration has its own, rather than sharing a type with the others: what they have in
// common today is --check, and a migration that asked for something else would be reaching
// into a struct built for its neighbours.
type ChecksumArgs struct {
	// Check writes nothing and reports what the migration would do, which is how a job
	// asks whether it has been run.
	Check bool
	// KeepFile leaves aqua-checksums.json where it is. Somebody else on the repository
	// may be on an aqua that still reads it, and this command can't know.
	KeepFile bool
}

// context is what one configuration file's migration works from.
type target struct {
	cfgFilePath string
	cfg         *aqua.Config
}

// targets is every configuration the directory has, read with its imports.
func (c *Controller) targets(logger *slog.Logger, wd, cfgFilePath string) ([]*target, error) {
	var out []*target
	for _, path := range c.configFinder.Finds(wd, cfgFilePath) {
		cfg := &aqua.Config{}
		if err := c.configReader.Read(logger, path, cfg); err != nil {
			return nil, err //nolint:wrapcheck
		}
		out = append(out, &target{cfgFilePath: path, cfg: cfg})
	}
	return out, nil
}
