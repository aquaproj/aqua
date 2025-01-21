package vacuum

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type PackageVacuumEntries []PackageVacuumEntry

type PackageVacuumEntry struct {
	PkgPath      []byte
	PackageEntry *PackageEntry
}

type PackageEntry struct {
	LastUsageTime time.Time
	Package       *Package
}

type Package struct {
	Name    string // Name of package (e.g. "cli/cli")
	Version string // Version of package (e.g. "v1.0.0")
	PkgPath string // Path to the install path without the rootDir/pkgs/ prefix
}

// Vacuum performs the vacuuming process if it is enabled.
func (vc *Controller) Vacuum(ctx context.Context, logE *logrus.Entry) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	return vc.vacuumExpiredPackages(ctx, logE)
}

// Close closes the dependencies of the Controller.
func (vc *Controller) Close(logE *logrus.Entry) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	logE.Debug("closing vacuum controller")
	if vc.db.storeQueue != nil {
		vc.db.storeQueue.close()
	}
	return vc.db.Close()
}

func (vc *Controller) TestKeepDBOpen() error {
	return vc.db.TestKeepDBOpen()
}

// StorePackage stores the given package if vacuum is enabled.
// If the package is nil, it logs a warning and skips storing the package.
func (vc *Controller) StorePackage(logE *logrus.Entry, pkg *config.Package, pkgPath string) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	if pkg == nil {
		logE.Warn("package is nil, skipping store package")
		return nil
	}
	return vc.handleAsyncStorePackage(logE, vc.getVacuumPackage(pkg, pkgPath))
}

// IsVacuumEnabled checks if the vacuum feature is enabled based on the configuration.
func (vc *Controller) IsVacuumEnabled(logE *logrus.Entry) bool {
	if vc.Param.VacuumDays <= 0 {
		logE.Debug("vacuum is disabled. AQUA_VACUUM_DAYS is not set or invalid.")
		return false
	}
	return true
}

// GetPackageLastUsed retrieves the last used time of a package. for testing purposes.
func (vc *Controller) GetPackageLastUsed(ctx context.Context, logE *logrus.Entry, pkgPath string) *time.Time {
	var lastUsedTime time.Time
	pkgEntry, _ := vc.db.Get(ctx, logE, pkgPath)
	if pkgEntry != nil {
		lastUsedTime = pkgEntry.LastUsageTime
	}
	return &lastUsedTime
}

// SetTimeStampPackage permit define a Timestamp for a package Manually. for testing purposes.
func (vc *Controller) SetTimestampPackage(ctx context.Context, logE *logrus.Entry, pkg *config.Package, pkgPath string, datetime time.Time) error {
	return vc.db.Store(ctx, logE, vc.getVacuumPackage(pkg, pkgPath), datetime)
}

// getVacuumPackage converts a config
func (vc *Controller) getVacuumPackage(configPkg *config.Package, pkgPath string) *Package {
	return &Package{
		Name:    configPkg.Package.Name,
		Version: configPkg.Package.Version,
		PkgPath: pkgPath,
	}
}

// handleAsyncStorePackage processes a list of configuration packages asynchronously.
func (vc *Controller) handleAsyncStorePackage(logE *logrus.Entry, vacuumPkg *Package) error {
	if vacuumPkg == nil {
		return errors.New("vacuumPkg is nil")
	}
	vc.db.StoreAsync(logE, vacuumPkg)
	return nil
}

const secondsInADay = 24 * 60 * 60

// isPackageExpired checks if a package is expired based on the vacuum configuration.
func (vc *Controller) isPackageExpired(pkg *PackageVacuumEntry) bool {
	threshold := vc.Param.VacuumDays * secondsInADay

	lastUsageTime := pkg.PackageEntry.LastUsageTime
	if lastUsageTime.Location() != time.UTC {
		lastUsageTime = lastUsageTime.In(time.UTC)
	}

	timeSinceLastUsage := time.Since(lastUsageTime).Seconds()
	return timeSinceLastUsage > float64(threshold)
}

// listExpiredPackages lists all packages that have expired based on the vacuum configuration.
func (vc *Controller) listExpiredPackages(ctx context.Context, logE *logrus.Entry) ([]*PackageVacuumEntry, error) {
	pkgs, err := vc.db.List(ctx, logE)
	if err != nil {
		return nil, err
	}

	var expired []*PackageVacuumEntry
	for _, pkg := range pkgs {
		if vc.isPackageExpired(pkg) {
			expired = append(expired, pkg)
		}
	}
	return expired, nil
}

// vacuumExpiredPackages performs cleanup of expired packages.
func (vc *Controller) vacuumExpiredPackages(ctx context.Context, logE *logrus.Entry) error {
	expiredPackages, err := vc.listExpiredPackages(ctx, logE)
	if err != nil {
		return err
	}

	if len(expiredPackages) == 0 {
		logE.Info("no expired packages to remove")
		return nil
	}

	successfulRemovals, errorsEncountered := vc.processExpiredPackages(logE, expiredPackages)

	if len(errorsEncountered) > 0 {
		var gErr error
		for _, err := range errorsEncountered {
			logerr.WithError(logE, err).Error("removing package path from system")
		}
		gErr = fmt.Errorf("total of %d errors encountered while removing package paths", len(errorsEncountered))
		return gErr
	}

	if len(successfulRemovals) > 0 {
		if err := vc.db.RemovePackages(ctx, logE, successfulRemovals); err != nil {
			return fmt.Errorf("remove packages from database: %w", err)
		}
	}

	return nil
}

// processExpiredPackages processes a list of expired package entries by removing their associated paths
// and generating a list of configuration packages to be removed from vacuum database.
func (vc *Controller) processExpiredPackages(logE *logrus.Entry, expired []*PackageVacuumEntry) ([]string, []error) {
	const batchSize = 10
	successKeys := make(chan string, len(expired))
	errCh := make(chan error, len(expired))

	var wg sync.WaitGroup
	for i := 0; i < len(expired); i += batchSize {
		end := i + batchSize
		if end > len(expired) {
			end = len(expired)
		}

		batch := make([]struct {
			pkgPath string
			pkgName string
			version string
		}, len(expired[i:end]))

		for j, entry := range expired[i:end] {
			batch[j].pkgPath = string(entry.PkgPath)
			batch[j].pkgName = entry.PackageEntry.Package.Name
			batch[j].version = entry.PackageEntry.Package.Version
		}

		wg.Add(1)
		go func(batch []struct {
			pkgPath string
			pkgName string
			version string
		},
		) {
			defer wg.Done()
			for _, entry := range batch {
				if err := vc.removePackageVersionPath(vc.Param, entry.pkgPath); err != nil {
					logerr.WithError(logE, err).WithField("pkg_path", entry.pkgPath).Error("removing path")
					errCh <- err
					continue
				}
				successKeys <- entry.pkgPath
			}
		}(batch)
	}

	wg.Wait()
	close(successKeys)
	close(errCh)

	pathsToRemove := make([]string, 0, len(expired))
	for pkgPath := range successKeys {
		pathsToRemove = append(pathsToRemove, pkgPath)
	}

	errors := make([]error, 0, len(expired))
	for err := range errCh {
		errors = append(errors, err)
	}

	return pathsToRemove, errors
}

// removePackageVersionPath removes the specified package version directory and its parent directory if it becomes empty.
func (vc *Controller) removePackageVersionPath(param *config.Param, path string) error {
	if err := vc.fs.RemoveAll(filepath.Join(param.RootDir, path)); err != nil {
		return fmt.Errorf("remove package version directories: %w", err)
	}
	return nil
}
