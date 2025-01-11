package vacuum

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/dustin/go-humanize"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"go.etcd.io/bbolt"
)

type DBAccessType string

const (
	dbFile         string       = "vacuum.db"
	bucketNamePkgs string       = "packages"
	View           DBAccessType = "view"
	Update         DBAccessType = "update"
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
func (vc *Controller) Vacuum(logE *logrus.Entry) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}
	return vc.vacuumExpiredPackages(logE)
}

// ListPackages lists the packages based on the provided arguments.
// If the expired flag is set to true, it lists the expired packages.
// Otherwise, it lists all packages.
func (vc *Controller) ListPackages(logE *logrus.Entry, expired bool, args ...string) error {
	if expired {
		return vc.handleListExpiredPackages(logE, args...)
	}
	return vc.handleListPackages(logE, args...)
}

// handleListPackages retrieves a list of packages and displays them using a fuzzy search.
func (vc *Controller) handleListPackages(logE *logrus.Entry, args ...string) error {
	pkgs, err := vc.listPackages(logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, pkgs, args...)
}

// handleListExpiredPackages handles the process of listing expired packages
// and displaying them using a fuzzy search.
func (vc *Controller) handleListExpiredPackages(logE *logrus.Entry, args ...string) error {
	expiredPkgs, err := vc.listExpiredPackages(logE)
	if err != nil {
		return err
	}
	return vc.displayPackagesFuzzy(logE, expiredPkgs, args...)
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

	vacuumPkg := vc.getVacuumPackage(pkg, pkgPath)

	return vc.handleAsyncStorePackage(logE, vacuumPkg)
}

// getVacuumPackage converts a config
func (vc *Controller) getVacuumPackage(configPkg *config.Package, pkgPath string) *Package {
	pkgPath = generatePackageKey(vc.Param.RootDir, pkgPath)
	return &Package{
		Type:    configPkg.PackageInfo.Type,
		Name:    configPkg.Package.Name,
		Version: configPkg.Package.Version,
		PkgPath: pkgPath,
	}
}

// generatePackageKey generates a package key based on the root directory and package path.
func generatePackageKey(rootDir string, pkgPath string) string {
	const splitParts = 2
	pkgPath = strings.SplitN(pkgPath, rootDir+"/pkgs/", splitParts)[1]
	return pkgPath
}

// handleAsyncStorePackage processes a list of configuration packages asynchronously.
func (vc *Controller) handleAsyncStorePackage(logE *logrus.Entry, vacuumPkg *Package) error {
	if vacuumPkg == nil {
		return errors.New("vacuumPkg is nil")
	}
	vc.storeQueue.enqueue(logE, vacuumPkg)
	return nil
}

// IsVacuumEnabled checks if the vacuum feature is enabled based on the configuration.
func (vc *Controller) IsVacuumEnabled(logE *logrus.Entry) bool {
	if vc.Param.VacuumDays <= 0 {
		logE.Debug("vacuum is disabled. AQUA_VACUUM_DAYS is not set or invalid.")
		return false
	}
	return true
}

// listExpiredPackages lists all packages that have expired based on the vacuum configuration.
func (vc *Controller) listExpiredPackages(logE *logrus.Entry) ([]*PackageVacuumEntry, error) {
	pkgs, err := vc.listPackages(logE)
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

// isPackageExpired checks if a package is expired based on the vacuum configuration.
func (vc *Controller) isPackageExpired(pkg *PackageVacuumEntry) bool {
	const secondsInADay = 24 * 60 * 60
	threshold := vc.Param.VacuumDays * secondsInADay

	lastUsageTime := pkg.PackageEntry.LastUsageTime
	if lastUsageTime.Location() != time.UTC {
		lastUsageTime = lastUsageTime.In(time.UTC)
	}

	timeSinceLastUsage := time.Since(lastUsageTime).Seconds()
	return timeSinceLastUsage > float64(threshold)
}

// listPackages lists all stored package entries.
func (vc *Controller) listPackages(logE *logrus.Entry) ([]*PackageVacuumEntry, error) {
	db, err := vc.getDB()
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}
	defer db.Close()

	var pkgs []*PackageVacuumEntry

	err = vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
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
	}, View)
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
		return fmt.Errorf("failed to display packages: %w", err)
	}

	return nil
}

// vacuumExpiredPackages performs cleanup of expired packages.
func (vc *Controller) vacuumExpiredPackages(logE *logrus.Entry) error {
	expiredPackages, err := vc.listExpiredPackages(logE)
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

	defer vc.Close(logE)
	if len(successfulRemovals) > 0 {
		if err := vc.removePackages(logE, successfulRemovals); err != nil {
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
			pkgType string
			pkgName string
			version string
		}, len(expired[i:end]))

		for j, entry := range expired[i:end] {
			batch[j].pkgPath = string(entry.PkgPath)
			batch[j].pkgType = entry.PackageEntry.Package.Type
			batch[j].pkgName = entry.PackageEntry.Package.Name
			batch[j].version = entry.PackageEntry.Package.Version
		}

		wg.Add(1)
		go func(batch []struct {
			pkgPath string
			pkgType string
			pkgName string
			version string
		},
		) {
			defer wg.Done()
			for _, entry := range batch {
				if err := vc.removePackageVersionPath(vc.Param, entry.pkgPath, entry.pkgType); err != nil {
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

// storePackageInternal stores package entries in the database.
func (vc *Controller) storePackageInternal(logE *logrus.Entry, pkg *Package, dateTime ...time.Time) error {
	lastUsedTime := time.Now()
	if len(dateTime) > 0 {
		lastUsedTime = dateTime[0]
	}
	return vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return errors.New("bucket not found")
		}
		logE.WithFields(logrus.Fields{
			"name":     pkg.Name,
			"version":  pkg.Version,
			"pkg_path": pkg.PkgPath,
		}).Debug("storing package in vacuum database")

		pkgKey := pkg.PkgPath
		pkgEntry := &PackageEntry{
			LastUsageTime: lastUsedTime,
			Package:       pkg,
		}

		data, err := encodePackageEntry(pkgEntry)
		if err != nil {
			logerr.WithError(logE, err).WithFields(
				logrus.Fields{
					"name":     pkg.Name,
					"version":  pkg.Version,
					"pkg_path": pkg.PkgPath,
				}).Error("encode package")
			return fmt.Errorf("encode package %s: %w", pkg.Name, err)
		}

		if err := b.Put([]byte(pkgKey), data); err != nil {
			logerr.WithError(logE, err).WithField("pkgKey", pkgKey).Error("store package in vacuum database")
			return fmt.Errorf("store package %s: %w", pkg.Name, err)
		}
		return nil
	}, Update)
}

// removePackages removes package entries from the database.
func (vc *Controller) removePackages(logE *logrus.Entry, pkgs []string) error {
	return vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
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
	}, Update)
}

// removePackageVersionPath removes the specified package version directory and its parent directory if it becomes empty.
func (vc *Controller) removePackageVersionPath(param *config.Param, path string, pkgType string) error {
	pkgsPath := filepath.Join(param.RootDir, "pkgs")
	pkgsRemoveLimitPath := filepath.Join(pkgsPath, pkgType)
	pkgVersionPath := filepath.Join(pkgsPath, path)
	if err := vc.fs.RemoveAll(pkgVersionPath); err != nil {
		return fmt.Errorf("remove package version directories: %w", err)
	}

	currentPath := filepath.Dir(pkgVersionPath)
	for currentPath != pkgsRemoveLimitPath {
		dirIsEmpty, err := afero.IsEmpty(vc.fs, currentPath)
		if err != nil {
			return fmt.Errorf("check if directory is empty: %w", err)
		}
		if !dirIsEmpty {
			break
		}
		if err := vc.fs.RemoveAll(currentPath); err != nil {
			return fmt.Errorf("remove package directories: %w", err)
		}
		currentPath = filepath.Dir(currentPath)
	}
	return nil
}

// encodePackageEntry encodes a PackageEntry into a JSON byte slice.
func encodePackageEntry(pkgEntry *PackageEntry) ([]byte, error) {
	data, err := json.Marshal(pkgEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package entry: %w", err)
	}
	return data, nil
}

// decodePackageEntry decodes a JSON byte slice into a PackageEntry.
func decodePackageEntry(data []byte) (*PackageEntry, error) {
	var pkgEntry PackageEntry
	if err := json.Unmarshal(data, &pkgEntry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package entry: %w", err)
	}
	return &pkgEntry, nil
}

// GetPackageLastUsed retrieves the last used time of a package. for testing purposes.
func (vc *Controller) GetPackageLastUsed(logE *logrus.Entry, pkgPath string) *time.Time {
	var lastUsedTime time.Time
	pkgEntry, _ := vc.retrievePackageEntry(logE, pkgPath)
	if pkgEntry != nil {
		lastUsedTime = pkgEntry.LastUsageTime
	}
	return &lastUsedTime
}

// SetTimeStampPackage permit define a Timestamp for a package Manually. for testing purposes.
func (vc *Controller) SetTimestampPackage(logE *logrus.Entry, pkg *config.Package, pkgPath string, datetime time.Time) error {
	vacuumPkg := vc.getVacuumPackage(pkg, pkgPath)
	return vc.storePackageInternal(logE, vacuumPkg, datetime)
}

// retrievePackageEntry retrieves a package entry from the database by key. for testing purposes.
func (vc *Controller) retrievePackageEntry(logE *logrus.Entry, key string) (*PackageEntry, error) {
	var pkgEntry *PackageEntry
	key = generatePackageKey(vc.Param.RootDir, key)
	err := vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
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
	}, View)
	return pkgEntry, err
}
