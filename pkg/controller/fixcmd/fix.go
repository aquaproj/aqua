package fixcmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

// standardRegistry is the registry the table of other names is for. A package that
// doesn't say which registry it comes from comes from this one.
const standardRegistry = "standard"

// Fix brings each aqua.yaml that applies to the directory up to date.
//
// No aqua.yaml means nothing to do, which isn't an error: the command is run from
// wherever the user happens to be, and a directory outside any aqua project has nothing
// to do.
func (c *Controller) Fix(ctx context.Context, logger *slog.Logger, param *config.Param, args *Args) error {
	upToDate := true
	for _, cfgFilePath := range c.configFinder.Finds(param.CWD, param.ConfigFilePath) {
		ok, err := c.fixFile(ctx, logger, cfgFilePath, args)
		if err != nil {
			return err
		}
		upToDate = upToDate && ok
	}
	if !upToDate {
		// Only --check gets here: without it the files were written.
		return errOutOfDate
	}
	return nil
}

// fixFile brings one configuration file and the files it imports up to date, and reports
// whether they already were.
func (c *Controller) fixFile(ctx context.Context, logger *slog.Logger, cfgFilePath string, args *Args) (bool, error) {
	logger = logger.With("config_file_path", cfgFilePath)

	cfg := &aqua.Config{}
	if err := c.configReader.Read(logger, cfgFilePath, cfg); err != nil {
		return false, err //nolint:wrapcheck
	}

	// The imports are read with it, so a package declared in one is renamed in the file
	// it is declared in. Nothing else knows which file that is.
	renames := c.renames(ctx, logger, cfg, cfgFilePath, args)
	if len(renames) == 0 {
		return true, nil
	}

	upToDate := true
	for path, rename := range renames {
		ok, err := c.rewrite(logger.With("path", path), path, rename, args)
		if err != nil {
			return false, err
		}
		upToDate = upToDate && ok
	}
	return upToDate, nil
}

// renames is the packages to rename, by the file each one is declared in.
//
// Reading the table is what this costs, and it is read once for the whole configuration
// rather than once per package: it holds every alias the registry has.
func (c *Controller) renames(ctx context.Context, logger *slog.Logger, cfg *aqua.Config, cfgFilePath string, args *Args) map[string]map[string]string {
	names := packageNames(cfg)
	if len(names) == 0 {
		return nil
	}
	table := c.names.NameTable(ctx, logger, args.Offline)
	if table == nil {
		// A registry with no table renames nothing. An older aqua-registry-g2, or a
		// mirror of one, simply doesn't have the file, and offline there may be no
		// copy of it yet.
		logger.Debug("the registry has no table of other names")
		return nil
	}

	renames := map[string]map[string]string{}
	for name, files := range names {
		to := table.Resolve(name)
		if to == name {
			continue
		}
		for _, file := range files {
			if file == "" {
				file = cfgFilePath
			}
			if renames[file] == nil {
				renames[file] = map[string]string{}
			}
			renames[file][name] = to
		}
	}
	return renames
}

// packageNames is the name of every package the table could have something to say
// about, and the files it is declared in.
//
// Only the standard registry's packages. A package from another registry is named the
// way that registry names it, and the table describes this one.
func packageNames(cfg *aqua.Config) map[string][]string {
	names := map[string][]string{}
	for _, pkg := range cfg.Packages {
		if pkg == nil || pkg.Name == "" {
			continue
		}
		if pkg.Registry != "" && pkg.Registry != standardRegistry {
			continue
		}
		// aqua.yaml writes a version either as part of the name or as a field of its
		// own, and what the table is asked about is the name.
		name, _, _ := strings.Cut(pkg.Name, "@")
		names[name] = append(names[name], pkg.FilePath)
	}
	return names
}

// rewrite renames the packages in one file, and reports whether it was already up to
// date.
func (c *Controller) rewrite(logger *slog.Logger, path string, renames map[string]string, args *Args) (bool, error) {
	for from, to := range renames {
		logger.Info("the registry calls the package something else now",
			"package_name", from, "renamed_to", to)
	}
	if args.Check {
		return false, nil
	}
	changed, err := renameInFile(logger, path, renames)
	if err != nil {
		return false, fmt.Errorf("rename the packages in %s: %w", path, err)
	}
	if !changed {
		// The names came from the configuration as it was read, imports resolved, so a
		// file that holds none of them is one whose packages were declared elsewhere.
		logger.Debug("nothing to rename in the file")
	}
	return true, nil
}
