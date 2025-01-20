package vacuum_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/vacuum"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/afero"
)

func TestVacuum(t *testing.T) { //nolint:funlen,maintidx,cyclop,gocognit,gocyclo
	t.Parallel()

	fs := afero.NewOsFs()

	t.Run("vacuum disabled", func(t *testing.T) {
		t.Parallel()
		logger, _ := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		// Setup
		param := &config.Param{
			RootDir: t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatalf("should return nil when vacuum is disabled: %v", err)
		}

		if err := controller.Vacuum(ctx, logE); err != nil {
			t.Fatalf("should return nil when vacuum is disabled: %v", err)
		}

		if err := controller.Close(logE); err != nil {
			t.Fatalf("should return nil when vacuum is disabled: %v", err)
		}
	})

	t.Run("vacuum bad configuration", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		logE.Logger.Level = logrus.DebugLevel
		testDir := t.TempDir()

		param := &config.Param{
			RootDir:    testDir,
			VacuumDays: -1,
		}
		controller := vacuum.New(context.Background(), param, fs)

		if err := controller.StorePackage(logE, nil, testDir); err != nil {
			t.Fatalf("should return nil when vacuum is disabled: %v", err)
		}
		if diff := cmp.Diff("vacuum is disabled. AQUA_VACUUM_DAYS is not set or invalid.", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
	})

	t.Run("ListPackages mode - empty database", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 30,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff("no packages to display", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
	})

	t.Run("StoreFailed", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 1, // Short expiration for testing
			RootDir:    t.TempDir(),
		}
		controller := vacuum.New(context.Background(), param, fs)

		numberPackagesToStore := 7
		pkgs := generateTestPackages(numberPackagesToStore, param.RootDir)

		// We force Keeping the DB open to simulate a failure in the async operation
		if err := controller.TestKeepDBOpen(); err != nil {
			t.Fatal(err)
		}

		hook.Reset()
		for _, pkg := range pkgs {
			if err := controller.StorePackage(logE, pkg.configPkg, pkg.pkgPath); err != nil {
				t.Fatal(err)
			}
		}

		// Wait for the async operations to complete
		if err := controller.Close(logE); err != nil {
			t.Fatal(err) // If AsyncStorePackage fails, Close should wait for the async operations to complete, but not return an error
		}

		time.Sleep(1 * time.Second) // Wait for ensure have time to get all logs
		expectedLogMessage := []string{
			"store package asynchronously",
			"retrying database operation",
			"store package asynchronously during shutdown",
		}
		var receivedMessages []string
		for _, entry := range hook.AllEntries() {
			receivedMessages = append(receivedMessages, entry.Message)
		}
		for _, entry := range expectedLogMessage {
			if !contains(receivedMessages, entry) {
				t.Errorf("Expected log message %q not found", entry)
			}
		}
	})

	t.Run("StorePackage and ListPackages workflow", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 30,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		numberPackagesToStore := 1
		pkgs := generateTestPackages(numberPackagesToStore, param.RootDir)

		// Store the package
		if err := controller.StorePackage(logE, pkgs[0].configPkg, pkgs[0].pkgPath); err != nil {
			t.Fatal(err)
		}
		// Close to ensure async operations are completed
		if err := controller.Close(logE); err != nil {
			t.Fatal(err)
		}

		// List packages - should contain our stored package
		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff("Test mode: Displaying packages", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(1, hook.LastEntry().Data["TotalPackages"]); diff != "" {
			t.Errorf("Unexpected total packages (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(0, hook.LastEntry().Data["TotalExpired"]); diff != "" {
			t.Errorf("Unexpected total expired (-want +got):\n%s", diff)
		}
		hook.Reset()

		// Verify package was stored correctly
		lastUsed := controller.GetPackageLastUsed(ctx, logE, pkgs[0].pkgPath)
		if lastUsed.IsZero() {
			t.Fatal("Package should have a last used time")
		}
	})

	t.Run("StoreMultiplePackages", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 30,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		numberPackagesToStore := 4
		pkgs := generateTestPackages(numberPackagesToStore, param.RootDir)

		// Store the package
		for _, pkg := range pkgs {
			err := controller.StorePackage(logE, pkg.configPkg, pkg.pkgPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Close to ensure async operations are completed
		if err := controller.Close(logE); err != nil {
			t.Fatal(err)
		}

		// List packages - should contain our stored package
		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff("Test mode: Displaying packages", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(4, hook.LastEntry().Data["TotalPackages"]); diff != "" {
			t.Errorf("Unexpected total packages (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(0, hook.LastEntry().Data["TotalExpired"]); diff != "" {
			t.Errorf("Unexpected total expired (-want +got):\n%s", diff)
		}
		hook.Reset()
	})

	t.Run("StoreNilPackage", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		testDir := t.TempDir()

		param := &config.Param{
			VacuumDays: 30,
			RootDir:    testDir,
		}
		controller := vacuum.New(context.Background(), param, fs)

		// Store the package
		if err := controller.StorePackage(logE, nil, testDir); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff("package is nil, skipping store package", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
	})

	t.Run("handleListExpiredPackages - no expired packages", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 30,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		if err := controller.ListPackages(ctx, logE, true, "test"); err != nil {
			t.Fatal(err) // Error if no package found
		}
		if diff := cmp.Diff("no packages to display", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}
	})
	t.Run("VacuumExpiredPackages workflow", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 1,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		numberPackagesToStore := 3
		numberPackagesToExpire := 1
		pkgs := generateTestPackages(numberPackagesToStore, param.RootDir)
		pkgPaths := make([]string, 0, numberPackagesToStore)

		// Create package paths and files
		for _, pkg := range pkgs {
			if err := fs.MkdirAll(pkg.pkgPath, 0o755); err != nil {
				t.Fatal(err)
			}

			// Create a test file in the package directory
			testFile := filepath.Join(pkg.pkgPath, "test.txt")
			if err := afero.WriteFile(fs, testFile, []byte("test content"), 0o644); err != nil {
				t.Fatal(err)
			}

			pkgPaths = append(pkgPaths, pkg.pkgPath)
		}

		// Store Multiple packages
		for _, pkg := range pkgs {
			if err := controller.StorePackage(logE, pkg.configPkg, pkg.pkgPath); err != nil {
				t.Fatal(err)
			}
		}

		// Call Close to ensure all async operations are completed
		if err := controller.Close(logE); err != nil {
			t.Fatal(err)
		}

		// Modify timestamp of one package to be expired
		oldTime := time.Now().Add(-48 * time.Hour) // 2 days old
		for _, pkg := range pkgs[:numberPackagesToExpire] {
			if err := controller.SetTimestampPackage(ctx, logE, pkg.configPkg, pkg.pkgPath, oldTime); err != nil {
				t.Fatal(err)
			}
		}

		// Check Packages after expiration
		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(numberPackagesToStore, hook.LastEntry().Data["TotalPackages"]); diff != "" {
			t.Errorf("Unexpected total packages (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(numberPackagesToExpire, hook.LastEntry().Data["TotalExpired"]); diff != "" {
			t.Errorf("Unexpected total expired (-want +got):\n%s", diff)
		}

		// List expired packages only
		if err := controller.ListPackages(ctx, logE, true, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(numberPackagesToExpire, hook.LastEntry().Data["TotalPackages"]); diff != "" {
			t.Errorf("Unexpected total packages (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(numberPackagesToExpire, hook.LastEntry().Data["TotalExpired"]); diff != "" {
			t.Errorf("Unexpected total expired (-want +got):\n%s", diff)
		}

		// Run vacuum
		if err := controller.Vacuum(ctx, logE); err != nil {
			t.Fatal(err)
		}

		// List expired packages
		if err := controller.ListPackages(ctx, logE, true, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff("no packages to display", hook.LastEntry().Message); diff != "" {
			t.Errorf("Unexpected log message (-want +got):\n%s", diff)
		}

		// Verify Package Paths was removed :
		for _, pkgPath := range pkgPaths[:numberPackagesToExpire] {
			exist, err := afero.Exists(fs, pkgPath)
			if err != nil {
				t.Fatal(err)
			}
			if exist {
				t.Fatal("Package directory should be removed after vacuum")
			}
		}

		// Modify timestamp of one package to be expired And lock DB to simulate a failure in the vacuum operation
		for _, pkg := range pkgs[:numberPackagesToExpire] {
			if err := controller.SetTimestampPackage(ctx, logE, pkg.configPkg, pkg.pkgPath, oldTime); err != nil {
				t.Fatal(err)
			}
		}

		// Keep Database open to simulate a failure in the vacuum operation
		if err := controller.TestKeepDBOpen(); err != nil {
			t.Fatal(err)
		}

		// Run vacuum
		if err := controller.Vacuum(ctx, logE); err == nil || !contains([]string{err.Error()}, "open database vacuum.db: timeout") {
			t.Fatalf("Expected timeout error, got %v", err)
		}
	})

	t.Run("TestVacuumWithoutExpiredPackages", func(t *testing.T) {
		t.Parallel()
		logger, hook := test.NewNullLogger()
		logE := logrus.NewEntry(logger)
		param := &config.Param{
			VacuumDays: 30,
			RootDir:    t.TempDir(),
		}
		ctx := context.Background()
		controller := vacuum.New(ctx, param, fs)

		// Store non-expired packages
		pkgs := generateTestPackages(3, param.RootDir)
		for _, pkg := range pkgs {
			err := controller.StorePackage(logE, pkg.configPkg, pkg.pkgPath)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Call Close to ensure all async operations are completed
		if err := controller.Close(logE); err != nil {
			t.Fatal(err)
		}

		// Run vacuum
		if err := controller.Vacuum(ctx, logE); err != nil {
			t.Fatal(err)
		}

		// Verify no packages were removed
		if err := controller.ListPackages(ctx, logE, false, "test"); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(3, hook.LastEntry().Data["TotalPackages"]); diff != "" {
			t.Errorf("Unexpected total packages (-want +got):\n%s", diff)
		}
	})
}

func contains(receivedMessages []string, entry string) bool {
	for _, msg := range receivedMessages {
		if msg == entry {
			return true
		}
	}
	return false
}

func TestMockVacuumController_StorePackage(t *testing.T) {
	t.Parallel()
	testDir := t.TempDir()

	param := &config.Param{
		RootDir: testDir,
	}

	logE := logrus.NewEntry(logrus.New())
	mockCtrl := vacuum.NewMockVacuumController()

	pkgs := generateTestPackages(2, param.RootDir)

	tests := []struct {
		name    string
		pkg     *config.Package
		pkgPath string
		wantErr bool
	}{
		{
			name:    "valid package",
			pkg:     pkgs[0].configPkg,
			pkgPath: pkgs[0].pkgPath,
			wantErr: false,
		},
		{
			name:    "nil package",
			pkg:     nil,
			pkgPath: pkgs[1].pkgPath,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := mockCtrl.StorePackage(logE, tt.pkg, tt.pkgPath); tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
			if err := mockCtrl.Vacuum(logE); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err := mockCtrl.Close(logE); err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNilVacuumController(t *testing.T) {
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())
	mockCtrl := &vacuum.NilVacuumController{}

	test := generateTestPackages(1, "/tmp")
	if err := mockCtrl.StorePackage(logE, test[0].configPkg, test[0].pkgPath); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if err := mockCtrl.Vacuum(logE); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if err := mockCtrl.Close(logE); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

type ConfigPackageWithPath struct {
	configPkg *config.Package
	pkgPath   string
}

func generateTestPackages(count int, rootDir string) []ConfigPackageWithPath {
	pkgs := make([]ConfigPackageWithPath, count)
	for i := range pkgs {
		pkgType := "github_release"
		pkgName := "cli/cli"
		version := "v2." + string(rune(i+'0')) + ".0"
		asset := "gh_2." + string(rune(i+'0')) + ".0_linux_amd64.tar.gz"
		pkgs[i] = ConfigPackageWithPath{
			configPkg: &config.Package{
				Package: &aqua.Package{
					Name:     pkgName,
					Version:  version,
					Registry: "standard",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      pkgType,
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     asset,
				},
			},
			pkgPath: filepath.Join(rootDir, "pkgs", pkgType, "github.com", pkgName, asset),
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
func benchmarkVacuumStorePackages(b *testing.B, pkgCount int) {
	b.Helper()
	logE := logrus.NewEntry(logrus.New())
	fs := afero.NewOsFs()

	syncf := b.TempDir()
	pkgs := generateTestPackages(pkgCount, syncf)
	syncParam := &config.Param{RootDir: syncf, VacuumDays: 5}

	b.Run("Sync", func(b *testing.B) {
		controller := vacuum.New(context.Background(), syncParam, fs)
		b.ResetTimer()
		for range b.N {
			for _, pkg := range pkgs {
				if err := controller.StorePackage(logE, pkg.configPkg, pkg.pkgPath); err != nil {
					b.Fatal(err)
				}
			}
			controller.Close(logE) // Close to ensure async operations are completed
		}
	})
}
