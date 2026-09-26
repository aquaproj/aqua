// Package fixcmd implements the aqua fix command.
//
// fix brings aqua.yaml up to date with the registry. A configuration goes out of date
// without being edited, because what it refers to moves: a package's repository is
// renamed, and the name in the file goes on working through an alias while naming
// something nobody uses.
//
// Nothing it does changes what the file means. The packages installed before and after a
// run are the same packages at the same versions, which is what makes it safe to run on
// every commit -- in a hook, or in the job that pushes fixes back to a pull request.
//
// That is the line between this and the other two commands that rewrite the file. 'aqua
// up' changes which version is installed: new code runs on the machine afterwards, so a
// person reads the diff. A migration changes the shape of a configuration for a new
// aqua, once, and carries decisions somebody has to make -- whether to keep
// aqua-checksums.json for whoever hasn't upgraded. Neither belongs in something a hook
// runs.
package fixcmd

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/g2"
)

// Args are the command's own options, as opposed to the global ones in config.Param.
type Args struct {
	// Check writes nothing and reports what is out of date, which is how a CI job asks
	// whether the configuration still says what the registry says.
	Check bool
	// Offline works from the table of other names already on disk instead of fetching
	// one. What can't be answered from it is left alone.
	Offline bool
}

// ConfigFinder lists the aqua.yaml files that apply to a directory.
type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

// ConfigReader reads an aqua.yaml, resolving its imports.
type ConfigReader interface {
	Read(logger *slog.Logger, configFilePath string, cfg *aqua.Config) error
}

// Names is the registry's record of what it calls a package.
//
// Only the standard registry has one. It is the registry aqua-registry-g2 holds, and
// the table sits beside its catalogue; another registry's aliases are its own business
// and are read from it at install time.
type Names interface {
	NameTable(ctx context.Context, logger *slog.Logger, offline bool) *g2.Aliases
}

// Controller brings configuration files up to date.
type Controller struct {
	configFinder ConfigFinder
	configReader ConfigReader
	names        Names
}

// New creates a Controller.
func New(configFinder ConfigFinder, configReader ConfigReader, names Names) *Controller {
	return &Controller{
		configFinder: configFinder,
		configReader: configReader,
		names:        names,
	}
}
