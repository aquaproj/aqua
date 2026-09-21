// Package lockupdate implements the aqua lock update command.
//
// The command builds aqua-lock.json from aqua.yaml. aqua.yaml says which package and
// which version; the lock file says what that version actually is on every
// environment, down to the checksum. Installing then needs no registry, which is what
// makes the lock file the thing aqua trusts rather than a cache of something else.
package lockupdate

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
)

// ConfigFinder lists the aqua.yaml files that apply to a directory.
type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

// ConfigReader reads an aqua.yaml, resolving its imports.
type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

// Resolver turns a package version into lock file entries, one per environment.
//
// Resolution has an order to it: the local registry cache, then aqua-registry-g2,
// then aqua-registry. They differ in where they read from, not in what they answer,
// so they share this interface and the controller tries them in turn.
type Resolver interface {
	Resolve(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]*lockfile.Package, error)
}

// Controller updates lock files.
type Controller struct {
	configFinder ConfigFinder
	configReader ConfigReader
	resolvers    []Resolver
}

// New creates a Controller.
func New(configFinder ConfigFinder, configReader ConfigReader, g2 Resolver) *Controller {
	return &Controller{
		configFinder: configFinder,
		configReader: configReader,
		resolvers:    []Resolver{g2},
	}
}
