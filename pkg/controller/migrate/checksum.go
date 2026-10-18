package migrate

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// errNotMigrated is what --check reports with, and what a configuration nothing was done
// to gets.
var errNotMigrated = errors.New("the configuration hasn't been migrated")

// Checksum takes out what aqua-checksums.json was for.
//
// A lock file records the checksum of every environment of every package a configuration
// asks for, so a repository with one has said the same thing twice: aqua-checksums.json
// is the same answer kept where install used to look for it. What the settings in
// aqua.yaml did -- turn the file on, insist every package be in it -- a lock file does by
// being the thing install reads.
//
// It is refused for a configuration whose lock file doesn't cover every package. The
// settings coming out is the checksum file stopping being read, and doing that while
// something isn't in the lock file would leave that package installed against nothing.
// 'aqua lock update' is what to run first.
//
// supported_envs stays. It says which environments a lock file records, and that is still
// a question: it is under checksum because that is where it was asked, and moving it is
// not this command's business.
func (c *Controller) Checksum(logger *slog.Logger, wd, cfgFilePath string, args *ChecksumArgs) error {
	targets, err := c.targets(logger, wd, cfgFilePath)
	if err != nil {
		return err
	}
	migrated := true
	for _, t := range targets {
		ok, err := c.checksumFile(logger, t, args)
		if err != nil {
			return err
		}
		migrated = migrated && ok
	}
	if !migrated {
		return errNotMigrated
	}
	return nil
}

// checksumFile migrates one configuration, and says whether it is migrated now.
func (c *Controller) checksumFile(logger *slog.Logger, t *target, args *ChecksumArgs) (bool, error) {
	logger = logger.With("config_file_path", t.cfgFilePath)
	dir := filepath.Dir(t.cfgFilePath)
	lockFilePath := filepath.Join(dir, lockfile.FileName)
	checksumPath := filepath.Join(dir, "aqua-checksums.json")

	settings, err := checksumSettings(t.cfgFilePath)
	if err != nil {
		return false, err
	}
	_, checksumErr := os.Stat(checksumPath)
	hasFile := checksumErr == nil
	if len(settings) == 0 && !hasFile {
		logger.Debug("there is nothing of aqua-checksums.json left to take out")
		return true, nil
	}

	if err := c.covered(logger, t, lockFilePath); err != nil {
		return false, err
	}

	for _, setting := range settings {
		logger.Info("the lock file says this, so the setting is what it was", "setting", setting)
	}
	if hasFile {
		logger.Info("the lock file holds the checksums, so the file is what it was", "path", checksumPath)
	}
	if args.Check {
		return false, nil
	}
	return true, c.takeOutChecksum(logger, t.cfgFilePath, checksumPath, settings, hasFile, args)
}

// takeOutChecksum writes the configuration without the settings and removes the file.
func (c *Controller) takeOutChecksum(logger *slog.Logger, cfgFilePath, checksumPath string, settings []string, hasFile bool, args *ChecksumArgs) error {
	if len(settings) > 0 {
		if err := removeChecksumSettings(cfgFilePath); err != nil {
			return err
		}
	}
	if !hasFile || args.KeepFile {
		return nil
	}
	if err := os.Remove(checksumPath); err != nil {
		// The settings are out, which is the half that decides what aqua reads. A file
		// nothing reads any more is worth reporting rather than failing for.
		slogerr.WithError(logger, err).Warn("remove aqua-checksums.json", "path", checksumPath)
		return nil
	}
	logger.Info("removed aqua-checksums.json", "path", checksumPath)
	return nil
}

// covered says every package the configuration asks for is in the lock file, for every
// environment the lock file has anything of.
//
// What is checked is that the package is there at all rather than that every environment
// is: a lock file records the environments the package supports, and which those are is
// the registry's answer rather than this command's.
func (c *Controller) covered(logger *slog.Logger, t *target, lockFilePath string) error {
	lf, err := lockfile.ReadFile(lockFilePath)
	if err != nil {
		return fmt.Errorf("read the lock file: %w", err)
	}
	if lf == nil {
		return fmt.Errorf("%w: %s", errNoLockFile, lockFilePath)
	}
	locked := make(map[string]struct{}, len(lf.Packages))
	for _, pkg := range lf.Packages {
		if pkg != nil {
			locked[pkg.Name+"\t"+pkg.Version] = struct{}{}
		}
	}

	missing := notLocked(t, locked)
	if len(missing) == 0 {
		return nil
	}
	for _, pkg := range missing {
		logger.Error("the lock file doesn't hold this package", "package", pkg)
	}
	return fmt.Errorf("%w: %d of them", errNotLocked, len(missing))
}

// notLocked is the packages the configuration asks for that the lock file has no entry for,
// as "<name>@<version>".
//
// A package with no version is left out of the answer: nothing can be looked up for it, and
// the lock file has nothing to say about it either.
func notLocked(t *target, locked map[string]struct{}) []string {
	var missing []string
	for _, pkg := range t.cfg.Packages {
		if pkg == nil || pkg.Name == "" {
			continue
		}
		name, version := pkg.Name, pkg.Version
		if version == "" {
			var found bool
			name, version, found = cutVersion(pkg.Name)
			if !found {
				continue
			}
		}
		if _, ok := locked[name+"\t"+version]; !ok {
			missing = append(missing, name+"@"+version)
		}
	}
	return missing
}
