package vacuum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/dustin/go-humanize"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
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
	Key          []byte
	PackageEntry *PackageEntry
}

type ConfigPackage struct {
	Type    string
	Name    string
	Version string
	PkgPath []string
}

type PackageEntry struct {
	LastUsageTime time.Time
	PkgPath       []string
}

type Mode string

const (
	ListPackages          Mode = "list-packages"
	ListExpiredPackages   Mode = "list-expired-packages"
	StorePackage          Mode = "store-package"
	StorePackages         Mode = "store-packages"
	AsyncStorePackage     Mode = "async-store-package"
	VacuumExpiredPackages Mode = "vacuum-expired-packages"
	Close                 Mode = "close"
)

// GetVacuumModeCLI returns the corresponding Mode based on the provided mode string.
// It supports the following modes:
// - "list-packages": returns ListPackages mode
// - "list-expired-packages": returns ListExpiredPackages mode
// - "vacuum-expired-packages": returns VacuumExpiredPackages mode
// If the provided mode string does not match any of the supported modes, it returns an error.
//
// Parameters:
//
//	mode (string): The mode string to be converted to a Mode.
//
// Returns:
//
//	(Mode, error): The corresponding Mode and nil if the mode string is valid, otherwise an empty Mode and an error.
func (vc *Controller) GetVacuumModeCLI(mode string) (Mode, error) {
	switch mode {
	case "list-packages":
		return ListPackages, nil
	case "list-expired-packages":
		return ListExpiredPackages, nil
	case "vacuum-expired-packages":
		return VacuumExpiredPackages, nil
	default:
		return "", errors.New("invalid vacuum mode")
	}
}

// Vacuum performs various vacuum operations based on the provided mode.
// Main function of vacuum controller.
func (vc *Controller) Vacuum(_ context.Context, logE *logrus.Entry, mode Mode, configPkg []*config.Package, args ...string) error {
	if !vc.IsVacuumEnabled(logE) {
		return nil
	}

	vacuumPkg := convertConfigPackages(configPkg)

	switch mode {
	case ListPackages:
		return vc.handleListPackages(logE, args...)
	case ListExpiredPackages:
		return vc.handleListExpiredPackages(logE, args...)
	case StorePackage:
		return vc.handleStorePackage(logE, vacuumPkg)
	case StorePackages:
		return vc.handleStorePackages(logE, vacuumPkg)
	case AsyncStorePackage:
		return vc.handleAsyncStorePackage(logE, vacuumPkg)
	case VacuumExpiredPackages:
		return vc.handleVacuumExpiredPackages(logE)
	case Close:
		return vc.close(logE)
	}

	return errors.New("invalid vacuum mode")
}

// convertConfigPackages converts a slice of config.Package pointers to a slice of ConfigPackage pointers.
func convertConfigPackages(configPkg []*config.Package) []*ConfigPackage {
	vacuumPkg := make([]*ConfigPackage, 0, len(configPkg))
	for _, pkg := range configPkg {
		vacuumPkg = append(vacuumPkg, vacuumConfigPackageFromConfigPackage(pkg))
	}
	return vacuumPkg
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

// handleStorePackage processes a list of configuration packages and stores the first package in the list.
func (vc *Controller) handleStorePackage(logE *logrus.Entry, vacuumPkg []*ConfigPackage) error {
	if len(vacuumPkg) < 1 {
		return errors.New("StorePackage requires at least one configPackage")
	}
	defer func() {
		if err := vc.close(logE); err != nil {
			logE.WithError(err).Error("Failed to close vacuum DB after storing package")
		}
	}()
	return vc.storePackageInternal(logE, []*ConfigPackage{vacuumPkg[0]})
}

// handleStorePackages processes a list of configuration packages and stores them.
func (vc *Controller) handleStorePackages(logE *logrus.Entry, vacuumPkg []*ConfigPackage) error {
	if len(vacuumPkg) < 1 {
		return errors.New("StorePackages requires at least one configPackage")
	}
	defer func() {
		if err := vc.close(logE); err != nil {
			logE.WithError(err).Error("Failed to close vacuum DB after storing multiple packages")
		}
	}()
	return vc.storePackageInternal(logE, vacuumPkg)
}

// handleAsyncStorePackage processes a list of configuration packages asynchronously.
func (vc *Controller) handleAsyncStorePackage(logE *logrus.Entry, vacuumPkg []*ConfigPackage) error {
	if len(vacuumPkg) < 1 {
		return errors.New("AsyncStorePackage requires at least one configPackage")
	}
	vc.storeQueue.enqueue(logE, vacuumPkg)
	return nil
}

// handleVacuumExpiredPackages handles the process of vacuuming expired packages.
func (vc *Controller) handleVacuumExpiredPackages(logE *logrus.Entry) error {
	return vc.vacuumExpiredPackages(logE)
}

// IsVacuumEnabled checks if the vacuum feature is enabled based on the configuration.
func (vc *Controller) IsVacuumEnabled(logE *logrus.Entry) bool {
	if vc.Param.VacuumDays == nil || *vc.Param.VacuumDays <= 0 {
		logE.Debug("Vacuum is disabled. AQUA_VACUUM_DAYS is not set or invalid.")
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
	threshold := int64(*vc.Param.VacuumDays) * secondsInADay
	return time.Since(pkg.PackageEntry.LastUsageTime).Seconds() > float64(threshold)
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
				logE.WithFields(logrus.Fields{
					"pkgKey": string(k),
				}).Warnf("Failed to decode entry: %v", err)
				return err
			}
			pkgs = append(pkgs, &PackageVacuumEntry{
				Key:          append([]byte{}, k...),
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
		logE.Info("No packages to display")
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
			pkgs[i].Key,
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
			parsedConfigPkg, err := generateConfigPackageFromKey(pkg.Key)
			if err != nil {
				return fmt.Sprintf("Failed to parse package key: %v", err)
			}
			return fmt.Sprintf(
				"Package Details:\n\n"+
					"%s \n"+
					"Type: %s\n"+
					"Package: %s\n"+
					"Version: %s\n\n"+
					"Last Used: %s\n"+
					"Last Used (exact): %s\n\n"+
					"PkgPath: %v\n",
				expiredString,
				parsedConfigPkg.Type,
				parsedConfigPkg.Name,
				parsedConfigPkg.Version,
				humanize.Time(pkg.PackageEntry.LastUsageTime),
				pkg.PackageEntry.LastUsageTime.Format("2024-12-31 15:04:05"),
				pkg.PackageEntry.PkgPath,
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
	expired, err := vc.listExpiredPackages(logE)
	if err != nil {
		return err
	}

	if len(expired) == 0 {
		return nil
	}

	successKeys, errCh := vc.processExpiredPackages(logE, expired)

	if len(errCh) > 0 {
		return errors.New("some packages could not be removed")
	}
	if len(successKeys) > 0 {
		if err := vc.removePackages(logE, successKeys); err != nil {
			return fmt.Errorf("failed to remove packages from database: %w", err)
		}
	}

	return nil
}

// processExpiredPackages processes a list of expired package entries by removing their associated paths
// and generating a list of configuration packages to be removed from vacuum database.
//
// Parameters:
//   - logE: A logrus.Entry used for logging errors and information.
//   - expired: A slice of PackageVacuumEntry representing the expired packages to be processed.
//
// Returns:
//   - A slice of ConfigPackage representing the packages that were successfully processed and need to be removed from the vacuum database.
//   - A slice of errors encountered during the processing of the expired packages.
func (vc *Controller) processExpiredPackages(logE *logrus.Entry, expired []*PackageVacuumEntry) ([]*ConfigPackage, []error) { //nolint:funlen
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
			key     string
			pkgPath []string
			version string
		}, len(expired[i:end]))

		for j, entry := range expired[i:end] {
			batch[j].key = string(entry.Key)
			batch[j].pkgPath = entry.PackageEntry.PkgPath
			batch[j].version = strings.Split(string(entry.Key), "@")[1]
		}

		wg.Add(1)
		go func(batch []struct {
			key     string
			pkgPath []string
			version string
		},
		) {
			defer wg.Done()
			for _, entry := range batch {
				for _, path := range entry.pkgPath {
					if err := vc.removePackageVersionPath(vc.Param, path, entry.version); err != nil {
						logE.WithField("expiredPackages", entry.key).WithError(err).Error("Error removing path")
						errCh <- err
						continue
					}
					successKeys <- entry.key
				}
			}
		}(batch)
	}

	wg.Wait()
	close(successKeys)
	close(errCh)

	ConfigPackageToRemove := make([]*ConfigPackage, 0, len(expired))
	for key := range successKeys {
		pkg, err := generateConfigPackageFromKey([]byte(key))
		if err != nil {
			logE.WithField("key", key).WithError(err).Error("Failed to generate package from key")
			continue
		}
		ConfigPackageToRemove = append(ConfigPackageToRemove, pkg)
	}

	errors := make([]error, 0, len(expired))
	for err := range errCh {
		errors = append(errors, err)
	}

	return ConfigPackageToRemove, errors
}

// storePackageInternal stores package entries in the database.
func (vc *Controller) storePackageInternal(logE *logrus.Entry, pkgs []*ConfigPackage, dateTime ...time.Time) error {
	lastUsedTime := time.Now()
	if len(dateTime) > 0 {
		lastUsedTime = dateTime[0]
	}
	return vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return errors.New("bucket not found")
		}

		for _, pkg := range pkgs {
			logE.WithFields(logrus.Fields{
				"name":    pkg.Name,
				"version": pkg.Version,
				"PkgPath": pkg.PkgPath,
			}).Debug("Storing package in vacuum database")
			pkgEntry := &PackageEntry{
				LastUsageTime: lastUsedTime,
				PkgPath:       pkg.PkgPath,
			}

			data, err := encodePackageEntry(pkgEntry)
			if err != nil {
				logE.WithFields(logrus.Fields{
					"name":    pkg.Name,
					"version": pkg.Version,
				}).WithError(err).Error("Failed to encode package")
				return fmt.Errorf("encode package %s: %w", pkg.Name, err)
			}

			if err := b.Put(generateKey(pkg), data); err != nil {
				logE.WithField("pkgKey", pkg).WithError(err).Error("Failed to store package in vacuum database")
				return fmt.Errorf("store package %s: %w", pkg.Name, err)
			}
		}
		return nil
	}, Update)
}

// removePackages removes package entries from the database.
func (vc *Controller) removePackages(logE *logrus.Entry, pkgs []*ConfigPackage) error {
	keys := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		keys = append(keys, string(generateKey(pkg)))
	}
	return vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return errors.New("bucket not found")
		}

		for _, key := range keys {
			if err := b.Delete([]byte(key)); err != nil {
				return fmt.Errorf("delete package %s: %w", key, err)
			}
		}
		return nil
	}, Update)
}

// removePackageVersionPath removes the specified package version directory and its parent directory if it becomes empty.
func (vc *Controller) removePackageVersionPath(param *config.Param, path string, version string) error {
	pkgVersionPath := filepath.Join(param.RootDir, "pkgs", path, version)
	if err := vc.fs.RemoveAll(pkgVersionPath); err != nil {
		return fmt.Errorf("remove package version directories: %w", err)
	}

	pkgPath := filepath.Join(param.RootDir, "pkgs", path)
	dirIsEmpty, err := afero.IsEmpty(vc.fs, pkgPath)
	if err != nil {
		return fmt.Errorf("check if the directory is empty: %w", err)
	}
	if dirIsEmpty {
		if err := vc.fs.RemoveAll(pkgPath); err != nil {
			return fmt.Errorf("remove package directories: %w", err)
		}
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

// retrievePackageEntry retrieves a package entry from the database by key.
func (vc *Controller) retrievePackageEntry(logE *logrus.Entry, key []byte) (*PackageEntry, error) {
	var pkgEntry *PackageEntry
	err := vc.withDBRetry(logE, func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketNamePkgs))
		if b == nil {
			return nil
		}
		value := b.Get(key)
		if value == nil {
			return nil
		}

		var err error
		pkgEntry, err = decodePackageEntry(value)
		return err
	}, View)
	return pkgEntry, err
}

// generateKey generates a unique key for a package.
func generateKey(pkg *ConfigPackage) []byte {
	return []byte(pkg.Type + "," + pkg.Name + "@" + pkg.Version)
}

// vacuumConfigPackageFromConfigPackage returns a ConfigPackage config from a config.Package.
func vacuumConfigPackageFromConfigPackage(pkg *config.Package) *ConfigPackage {
	pkgPathsMap := pkg.PackageInfo.PkgPaths()
	PkgPath := make([]string, 0, len(pkgPathsMap))
	for k := range pkgPathsMap {
		PkgPath = append(PkgPath, k)
	}

	return &ConfigPackage{
		Type:    pkg.PackageInfo.Type,
		Name:    pkg.Package.Name,
		Version: pkg.Package.Version,
		PkgPath: PkgPath,
	}
}

// generateConfigPackageFromKey return a minimal package config from a key.
func generateConfigPackageFromKey(key []byte) (*ConfigPackage, error) {
	if len(key) == 0 {
		return nil, errors.New("empty key")
	}
	pattern := `^(?P<PackageInfo_Type>[^,]+),(?P<Package_Name>[^@]+)@(?P<Package_Version>.+)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(string(key))

	if match == nil {
		return nil, fmt.Errorf("key %s does not match the pattern %s", key, pattern)
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	packageInfoType := result["PackageInfo_Type"]
	packageName := result["Package_Name"]
	packageVersion := result["Package_Version"]

	pkg := &ConfigPackage{
		Type:    packageInfoType,
		Name:    packageName,
		Version: packageVersion,
	}
	return pkg, nil
}
