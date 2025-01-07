package vacuum_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/vacuum"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ListPackages          string = "list-packages"
	ListExpiredPackages   string = "list-expired-packages"
	StorePackage          string = "store-package"
	StorePackages         string = "store-packages"
	AsyncStorePackage     string = "async-store-package"
	VacuumExpiredPackages string = "vacuum-expired-packages"
	Close                 string = "close"
)

func TestVacuum(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()
	// Setup common test fixtures
	logger, hook := test.NewNullLogger()
	logE := logrus.NewEntry(logger)

	ctx := context.Background()
	fs := afero.NewOsFs()

	// Create temp directory for tests
	tempTestDir, err := afero.TempDir(fs, "/tmp", "vacuum_test")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := fs.RemoveAll(tempTestDir)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("vacuum disabled", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_disabled")
		require.NoError(t, err)
		// Setup
		param := &config.Param{
			VacuumDays: nil,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil)

		// Assert
		require.NoError(t, err, "Should return nil when vacuum is disabled")
	})

	t.Run("invalid mode", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_invalid_mode")
		require.NoError(t, err)
		// Setup
		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test
		err = controller.Vacuum(ctx, logE, vacuum.Mode("invalid"), nil)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid vacuum mode")
	})

	t.Run("ListPackages mode - empty database", func(t *testing.T) {
		t.Parallel()
		// Setup - use a new temp directory for this test
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_list_test")
		require.NoError(t, err)

		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")

		// Assert
		require.NoError(t, err) // Should succeed with empty database
		assert.Equal(t, "No packages to display", hook.LastEntry().Message)
	})

	t.Run("AsyncFailed", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_async_failed")
		require.NoError(t, err)

		days := 1 // Short expiration for testing
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		numberPackagesToStore := 4
		pkgs := generateTestPackages(numberPackagesToStore)

		// We force Keeping the DB open to simulate a failure in the async operation
		if err := controller.TestKeepDBOpen(); err != nil {
			t.Fatal(err)
		}
		hook.Reset()
		for _, pkg := range pkgs {
			err := controller.Vacuum(context.Background(), logE, vacuum.Mode(AsyncStorePackage), []*config.Package{pkg})
			require.NoError(t, err)
		}

		// Wait for the async operations to complete
		err = controller.Vacuum(context.Background(), logE, vacuum.Mode(Close), nil)
		require.NoError(t, err) // If AsyncStorePackage fails, Close should wait for the async operations to complete, but not return an error

		expectedLogMessage := []string{
			"Failed to store package asynchronously",
			"Retrying database operation",
			"Failed to store package asynchronously during shutdown",
		}
		var receivedMessages []string
		for _, entry := range hook.AllEntries() {
			receivedMessages = append(receivedMessages, entry.Message)
		}
		for _, entry := range expectedLogMessage {
			assert.Contains(t, receivedMessages, entry)
		}
	})

	t.Run("StorePackage and ListPackages workflow", func(t *testing.T) {
		t.Parallel()
		// Setup - use a new temp directory for this test
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_store_test")
		require.NoError(t, err)

		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		numberPackagesToStore := 1
		pkgs := generateTestPackages(numberPackagesToStore)

		// Store the package
		err = controller.Vacuum(ctx, logE, vacuum.Mode(StorePackage), pkgs)
		require.NoError(t, err)

		// List packages - should contain our stored package
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, "Test mode: Displaying packages", hook.LastEntry().Message)
		assert.Equal(t, 1, hook.LastEntry().Data["TotalPackages"])
		assert.Equal(t, 0, hook.LastEntry().Data["TotalExpired"])
		hook.Reset()

		// Verify package was stored correctly
		lastUsed := controller.GetPackageLastUsed(logE, pkgs[0])
		assert.False(t, lastUsed.IsZero(), "Package should have a last used time")
	})

	t.Run("StorePackages", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_storePackages_test")
		require.NoError(t, err)

		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		numberPackagesToStore := 4
		pkgs := generateTestPackages(numberPackagesToStore)

		// Store the package
		err = controller.Vacuum(ctx, logE, vacuum.Mode(StorePackages), pkgs)
		require.NoError(t, err)

		// List packages - should contain our stored package
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, "Test mode: Displaying packages", hook.LastEntry().Message)
		assert.Equal(t, 4, hook.LastEntry().Data["TotalPackages"])
		assert.Equal(t, 0, hook.LastEntry().Data["TotalExpired"])
		hook.Reset()
	})

	t.Run("GetVacuumModeCLI valid modes", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_getVacuumModeCLI_test")
		require.NoError(t, err)
		param := &config.Param{
			RootDir: testDir,
		}
		controller := vacuum.New(param, fs)

		// Test valid modes
		modes := map[string]vacuum.Mode{
			"list-packages":           vacuum.ListPackages,
			"list-expired-packages":   vacuum.ListExpiredPackages,
			"vacuum-expired-packages": vacuum.VacuumExpiredPackages,
		}

		for modeStr, expectedMode := range modes {
			mode, err := controller.GetVacuumModeCLI(modeStr)
			require.NoError(t, err)
			assert.Equal(t, expectedMode, mode)
		}
	})

	t.Run("GetVacuumModeCLI invalid mode", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_getVacuumModeCLI_invalid_test")
		require.NoError(t, err)
		param := &config.Param{
			RootDir: testDir,
		}
		controller := vacuum.New(param, fs)

		// Test invalid mode
		mode, err := controller.GetVacuumModeCLI("invalid-mode")
		require.Error(t, err)
		assert.Equal(t, vacuum.Mode(""), mode)
		assert.Contains(t, err.Error(), "invalid vacuum mode")
	})

	t.Run("handleListExpiredPackages - no expired packages", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_handle_list_expired")
		require.NoError(t, err)

		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListExpiredPackages), nil, "test")

		// Assert
		require.NoError(t, err) // Error if no package found
		assert.Equal(t, "No packages to display", hook.LastEntry().Message)
	})

	t.Run("handleStorePackage - invalid package count", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_handle_store_package")
		require.NoError(t, err)
		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test with no packages
		err = controller.Vacuum(ctx, logE, vacuum.Mode(StorePackage), nil)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "StorePackage requires at least one configPackage")
	})

	t.Run("handleStorePackages - invalid package count", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_handle_store_packages")
		require.NoError(t, err)
		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test with no packages
		err = controller.Vacuum(ctx, logE, vacuum.Mode(StorePackages), nil)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "StorePackages requires at least one configPackage")
	})

	t.Run("handleAsyncStorePackage - invalid package count", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_handle_async_store_package")
		require.NoError(t, err)
		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		// Test with no packages
		err = controller.Vacuum(ctx, logE, vacuum.Mode(AsyncStorePackage), nil)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AsyncStorePackage requires at least one configPackage")
	})

	t.Run("VacuumExpiredPackages workflow", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(fs, tempTestDir, "vacuum_expire_test")
		require.NoError(t, err)
		defer func() {
			hook.Reset()
			fs.RemoveAll(testDir) //nolint:errcheck
		}()

		days := 1 // Short expiration for testing
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, fs)

		numberPackagesToStore := 3
		numberPackagesToExpire := 1
		pkgs := generateTestPackages(numberPackagesToStore)
		pkgPaths := make([]string, 0, len(pkgs))

		// Create package paths and files
		for _, pkg := range pkgs {
			pkgPath := filepath.Join(testDir, "pkgs", "github_release/github.com/"+pkg.Package.Name, pkg.Package.Version)
			err = fs.MkdirAll(pkgPath, 0o755)
			require.NoError(t, err)

			// Create a test file in the package directory
			testFile := filepath.Join(pkgPath, "test.txt")
			err = afero.WriteFile(fs, testFile, []byte("test content"), 0o644)
			require.NoError(t, err)

			pkgPaths = append(pkgPaths, pkgPath)
		}

		// Store Multiple packages
		err = controller.Vacuum(ctx, logE, vacuum.Mode(AsyncStorePackage), pkgs)
		require.NoError(t, err)

		// Call Close to ensure all async operations are completed
		err = controller.Vacuum(ctx, logE, vacuum.Mode(Close), nil)
		require.NoError(t, err)

		// Modify timestamp of one package to be expired
		oldTime := time.Now().Add(-48 * time.Hour) // 2 days old
		err = controller.SetTimestampPackages(logE, pkgs[0:numberPackagesToExpire], oldTime)
		require.NoError(t, err)

		// Check Packages after expiration
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, numberPackagesToStore, hook.LastEntry().Data["TotalPackages"])
		assert.Equal(t, numberPackagesToExpire, hook.LastEntry().Data["TotalExpired"])

		// List expired packages only
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListExpiredPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, numberPackagesToExpire, hook.LastEntry().Data["TotalPackages"])
		assert.Equal(t, numberPackagesToExpire, hook.LastEntry().Data["TotalExpired"])

		// Run vacuum
		err = controller.Vacuum(ctx, logE, vacuum.Mode(VacuumExpiredPackages), nil)
		require.NoError(t, err)

		// List expired packages
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, numberPackagesToStore-numberPackagesToExpire, hook.LastEntry().Data["TotalPackages"])
		assert.Equal(t, 0, hook.LastEntry().Data["TotalExpired"])

		// Verify Package Paths was removed :
		for _, pkgPath := range pkgPaths[:numberPackagesToExpire] {
			exist, err := afero.Exists(fs, pkgPath)
			require.NoError(t, err)
			assert.False(t, exist, "Package directory should be removed after vacuum")
		}

		// Modify timestamp of one package to be expired And lock DB to simulate a failure in the vacuum operation
		err = controller.SetTimestampPackages(logE, pkgs[numberPackagesToExpire:], oldTime)
		require.NoError(t, err)

		// Keep Database open to simulate a failure in the vacuum operation
		err = controller.TestKeepDBOpen()
		require.NoError(t, err)

		// Run vacuum
		err = controller.Vacuum(ctx, logE, vacuum.Mode(VacuumExpiredPackages), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open database vacuum.db: timeout")
	})

	t.Run("TestVacuumWithoutExpiredPackages", func(t *testing.T) {
		t.Parallel()
		testDir, err := afero.TempDir(afero.NewOsFs(), "", "vacuum_no_expired")
		require.NoError(t, err)

		days := 30
		param := &config.Param{
			VacuumDays: &days,
			RootDir:    testDir,
		}
		controller := vacuum.New(param, afero.NewOsFs())

		// Store non-expired packages
		pkgs := generateTestPackages(3)
		err = controller.Vacuum(ctx, logE, vacuum.Mode(StorePackages), pkgs)
		require.NoError(t, err)

		// Run vacuum
		err = controller.Vacuum(ctx, logE, vacuum.Mode(VacuumExpiredPackages), nil)
		require.NoError(t, err)

		// Verify no packages were removed
		err = controller.Vacuum(ctx, logE, vacuum.Mode(ListPackages), nil, "test")
		require.NoError(t, err)
		assert.Equal(t, 3, hook.LastEntry().Data["TotalPackages"])
	})
}

func generateTestPackages(count int) []*config.Package {
	pkgs := make([]*config.Package, count)
	for i := range count {
		pkgs[i] = &config.Package{
			Package: &aqua.Package{
				Name:     "cli/cli",
				Version:  "v2." + string(rune(i+'0')) + ".0",
				Registry: "standard",
			},
			PackageInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "cli",
				RepoName:  "cli",
				Asset:     "gh_2." + string(rune(i+'0')) + ".0_linux_amd64.tar.gz",
			},
		}
	}
	return pkgs
}

// BenchmarkVacuum_StorePackages benchmarks the performance of storing packages.
func BenchmarkVacuum_StorePackages(b *testing.B) {
	benchmarkVacuumStorePackages(b, 100)
}

// BenchmarkVacuum_OnlyOneStorePackage benchmarks the performance of storing only one package.
func BenchmarkVacuum_OnlyOneStorePackage(b *testing.B) {
	benchmarkVacuumStorePackages(b, 1)
}

// benchmarkVacuumStorePackages is a helper function to benchmark the performance of storing packages.
func benchmarkVacuumStorePackages(b *testing.B, pkgCount int) { //nolint:cyclop,funlen,gocognit
	b.Helper()
	pkgs, logE, syncParam, syncMultipleParam, asyncParam, asyncMultipleParam, fs, syncf, syncMultiplef, asyncf, asyncMultiplef := setupBenchmark(b, pkgCount)
	defer func() {
		if err := fs.RemoveAll(syncf); err != nil {
			b.Fatal(err)
		}
		if err := fs.RemoveAll(syncMultiplef); err != nil {
			b.Fatal(err)
		}
		if err := fs.RemoveAll(asyncf); err != nil {
			b.Fatal(err)
		}
		if err := fs.RemoveAll(asyncMultiplef); err != nil {
			b.Fatal(err)
		}
	}()

	b.Run("Sync", func(b *testing.B) {
		controller := vacuum.New(syncParam, fs)
		b.ResetTimer()
		for range b.N {
			for _, pkg := range pkgs {
				if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(StorePackage), []*config.Package{pkg}); err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("SyncMultipleSameTime", func(b *testing.B) {
		controller := vacuum.New(syncMultipleParam, fs)
		b.ResetTimer()
		for range b.N {
			if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(StorePackage), pkgs); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Async", func(b *testing.B) {
		controller := vacuum.New(asyncParam, fs)
		b.ResetTimer()
		for range b.N {
			runtime.MemProfileRate = 1
			for _, pkg := range pkgs {
				if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(AsyncStorePackage), []*config.Package{pkg}); err != nil {
					b.Fatal(err)
				}
			}
			if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(Close), nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("AsyncMultiple", func(b *testing.B) {
		controller := vacuum.New(asyncMultipleParam, fs)
		b.ResetTimer()
		for range b.N {
			if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(AsyncStorePackage), pkgs); err != nil {
				b.Fatal(err)
			}
			if err := controller.Vacuum(context.Background(), logE, vacuum.Mode(Close), nil); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func setupBenchmark(b *testing.B, pkgCount int) ([]*config.Package, *logrus.Entry, *config.Param, *config.Param, *config.Param, *config.Param, afero.Fs, string, string, string, string) {
	b.Helper()
	pkgs := generateTestPackages(pkgCount)
	vacuumDays := 5
	logE := logrus.NewEntry(logrus.New())
	fs := afero.NewOsFs()

	// Benchmark sync configuration
	syncf, errf := afero.TempDir(fs, "/tmp", "vacuum_test_sync")
	if errf != nil {
		b.Fatal(errf)
	}
	syncParam := &config.Param{RootDir: syncf, VacuumDays: &vacuumDays}

	// Benchmark SyncMultipleSameTime configuration
	syncMultiplef, errf := afero.TempDir(fs, "/tmp", "vacuum_test_multiple")
	if errf != nil {
		b.Fatal(errf)
	}
	syncMultipleParam := &config.Param{RootDir: syncMultiplef, VacuumDays: &vacuumDays}

	// Benchmark async configuration
	asyncf, errf := afero.TempDir(fs, "/tmp", "vacuum_test_async")
	if errf != nil {
		b.Fatal(errf)
	}
	asyncParam := &config.Param{RootDir: asyncf, VacuumDays: &vacuumDays}

	// Benchmark async multiple configuration
	asyncMultiplef, errf := afero.TempDir(fs, "/tmp", "vacuum_test_async_multiple")
	if errf != nil {
		b.Fatal(errf)
	}
	asyncMultipleParam := &config.Param{RootDir: asyncMultiplef, VacuumDays: &vacuumDays}

	return pkgs, logE, syncParam, syncMultipleParam, asyncParam, asyncMultipleParam, fs, syncf, syncMultiplef, asyncf, asyncMultiplef
}
