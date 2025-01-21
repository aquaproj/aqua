package vacuum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/dustin/go-humanize"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"go.etcd.io/bbolt"
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
	Type    string // Type of package (e.g. "github_release")
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

// ListPackages lists the packages based on the provided arguments.
// If the expired flag is set to true, it lists the expired packages.
// Otherwise, it lists all packages.
func (vc *Controller) ListPackages(ctx context.Context, logE *logrus.Entry, expired bool, args ...string) error {
	if expired {
		return vc.handleListExpiredPackages(ctx, logE, args...)
	}
	return vc.handleListPackages(ctx, logE, args...)
}

// Close closes the dependencies of the Controller.
func (vc *Controller) Close(logE *logrus.Entry) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	logE.Debug("closing vacuum controller")
	if vc.d.storeQueue != nil {
		vc.d.storeQueue.close()
	}
	return vc.d.Close()
}

func (vc *Controller) TestKeepDBOpen() error {
	return vc.d.TestKeepDBOpen()
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
	pkgEntry, _ := vc.retrievePackageEntry(ctx, logE, pkgPath)
	if pkgEntry != nil {
		lastUsedTime = pkgEntry.LastUsageTime
	}
	return &lastUsedTime
}

// SetTimeStampPackage permit define a Timestamp for a package Manually. for testing purposes.
func (vc *Controller) SetTimestampPackage(ctx context.Context, logE *logrus.Entry, pkg *config.Package, pkgPath string, datetime time.Time) error {
	return vc.d.Store(ctx, logE, vc.getVacuumPackage(pkg, pkgPath), datetime)
}

// handleListPackages retrieves a list of packages and displays them using a fuzzy search.
func (vc *Controller) handleListPackages(ctx context.Context, logE *logrus.Entry, args ...string) error {
	pkgs, err := vc.listPackages(ctx, logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, pkgs, args...)
}

// handleListExpiredPackages handles the process of listing expired packages
// and displaying them using a fuzzy search.
func (vc *Controller) handleListExpiredPackages(ctx context.Context, logE *logrus.Entry, args ...string) error {
	expiredPkgs, err := vc.listExpiredPackages(ctx, logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, expiredPkgs, args...)
}

// getVacuumPackage converts a config
func (vc *Controller) getVacuumPackage(configPkg *config.Package, pkgPath string) *Package {
	return &Package{
		Type:    configPkg.PackageInfo.Type,
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
	vc.d.storeQueue.enqueue(logE, vacuumPkg)
	return nil
}

// listExpiredPackages lists all packages that have expired based on the vacuum configuration.
func (vc *Controller) listExpiredPackages(ctx context.Context, logE *logrus.Entry) ([]*PackageVacuumEntry, error) {
	pkgs, err := vc.listPackages(ctx, logE)
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

// listPackages lists all stored package entries.
func (vc *Controller) listPackages(ctx context.Context, logE *logrus.Entry) ([]*PackageVacuumEntry, error) {
	db, err := vc.d.getDB()
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}

	var pkgs []*PackageVacuumEntry

	err = vc.d.view(ctx, logE, func(tx *bbolt.Tx) error {
		b := vc.d.Bucket(tx)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, value []byte) error {
			pkgEntry, err := decodePackageEntry(value)
			if err != nil {
				logerr.WithError(logE, err).WithField("pkg_key", string(k)).Warn("unable to decode entry")
				return err
			}
			pkgs = append(pkgs, &PackageVacuumEntry{
				PkgPath:      append([]byte{}, k...),
				PackageEntry: pkgEntry,
			})
			return nil
		})
	})
	return pkgs, err
}

func (vc *Controller) displayPackagesFuzzyTest(logE *logrus.Entry, pkgs []*PackageVacuumEntry) error {
	var pkgInformations struct {
		TotalPackages int
		TotalExpired  int
	}
	for _, pkg := range pkgs {
		if vc.isPackageExpired(pkg) {
			pkgInformations.TotalExpired++
		}
		pkgInformations.TotalPackages++
	}
	// Display log entry with informations for testing purposes
	logE.WithFields(logrus.Fields{
		"TotalPackages": pkgInformations.TotalPackages,
		"TotalExpired":  pkgInformations.TotalExpired,
	}).Info("Test mode: Displaying packages")
	return nil
}

func (vc *Controller) displayPackagesFuzzy(logE *logrus.Entry, pkgs []*PackageVacuumEntry, args ...string) error {
	if len(pkgs) == 0 {
		logE.Info("no packages to display")
		return nil
	}
	if len(args) > 0 && args[0] == "test" {
		return vc.displayPackagesFuzzyTest(logE, pkgs)
	}
	return vc.displayPackagesFuzzyInteractive(pkgs)
}

func (vc *Controller) displayPackagesFuzzyInteractive(pkgs []*PackageVacuumEntry) error {
	_, err := fuzzyfinder.Find(pkgs, func(i int) string {
		var expiredString string
		if vc.isPackageExpired(pkgs[i]) {
			expiredString = "⌛ "
		}

		return fmt.Sprintf("%s%s [%s]",
			expiredString,
			pkgs[i].PkgPath,
			humanize.Time(pkgs[i].PackageEntry.LastUsageTime),
		)
	},
		fuzzyfinder.WithPreviewWindow(func(i, _, _ int) string {
			if i == -1 {
				return "No package selected"
			}
			pkg := pkgs[i]
			var expiredString string
			if vc.isPackageExpired(pkg) {
				expiredString = "Expired ⌛"
			}
			return fmt.Sprintf(
				"Package Details:\n\n"+
					"%s \n"+
					"Type: %s\n"+
					"Package: %s\n"+
					"Version: %s\n\n"+
					"Last Used: %s\n"+
					"Last Used (exact): %s\n\n",
				expiredString,
				pkg.PackageEntry.Package.Type,
				pkg.PackageEntry.Package.Name,
				pkg.PackageEntry.Package.Version,
				humanize.Time(pkg.PackageEntry.LastUsageTime),
				pkg.PackageEntry.LastUsageTime.Format("2006-01-02 15:04:05"),
			)
		}),
		fuzzyfinder.WithHeader("Navigate through packages to display details"),
	)
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return fmt.Errorf("display packages: %w", err)
	}

	return nil
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
		if err := vc.removePackages(ctx, logE, successfulRemovals); err != nil {
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

// removePackages removes package entries from the database.
func (vc *Controller) removePackages(ctx context.Context, logE *logrus.Entry, pkgs []string) error {
	return vc.d.update(ctx, logE, func(tx *bbolt.Tx) error {
		b := vc.d.Bucket(tx)
		if b == nil {
			return errors.New("bucket not found")
		}

		for _, key := range pkgs {
			if err := b.Delete([]byte(key)); err != nil {
				return fmt.Errorf("delete package %s: %w", key, err)
			}
			logE.WithField("pkgKey", key).Info("removed package from vacuum database")
		}
		return nil
	})
}

// removePackageVersionPath removes the specified package version directory and its parent directory if it becomes empty.
func (vc *Controller) removePackageVersionPath(param *config.Param, path string) error {
	if err := vc.fs.RemoveAll(filepath.Join(param.RootDir, path)); err != nil {
		return fmt.Errorf("remove package version directories: %w", err)
	}
	return nil
}

// encodePackageEntry encodes a PackageEntry into a JSON byte slice.
func encodePackageEntry(pkgEntry *PackageEntry) ([]byte, error) {
	data, err := json.Marshal(pkgEntry)
	if err != nil {
		return nil, fmt.Errorf("marshal package entry: %w", err)
	}
	return data, nil
}

// decodePackageEntry decodes a JSON byte slice into a PackageEntry.
func decodePackageEntry(data []byte) (*PackageEntry, error) {
	var pkgEntry PackageEntry
	if err := json.Unmarshal(data, &pkgEntry); err != nil {
		return nil, fmt.Errorf("unmarshal package entry: %w", err)
	}
	return &pkgEntry, nil
}

// retrievePackageEntry retrieves a package entry from the database by key. for testing purposes.
func (vc *Controller) retrievePackageEntry(ctx context.Context, logE *logrus.Entry, key string) (*PackageEntry, error) {
	var pkgEntry *PackageEntry
	err := vc.d.view(ctx, logE, func(tx *bbolt.Tx) error {
		b := vc.d.Bucket(tx)
		if b == nil {
			return nil
		}
		value := b.Get([]byte(key))
		if value == nil {
			return nil
		}

		var err error
		pkgEntry, err = decodePackageEntry(value)
		return err
	})
	return pkgEntry, err
}
