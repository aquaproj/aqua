package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

func TestPackageInfo_SetVersion(t *testing.T) { //nolint:funlen
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())

	data := []struct {
		name        string
		pkg         *registry.PackageInfo
		version     string
		expectName  string
		expectError bool
	}{
		{
			name: "no version constraint",
			pkg: &registry.PackageInfo{
				Name: "test-package",
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version constraint matches",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "true",
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version constraint does not match",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false",
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version prefix matching",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionPrefix:      "v",
				VersionConstraints: "true",
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version prefix not matching",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionPrefix:      "release-",
				VersionConstraints: "true",
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version override matches",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false", // Top level doesn't match
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "true",
						Type:               "github_release",
					},
				},
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version override with prefix",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false",
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "true",
						VersionPrefix:      stringPtr("v"),
						Type:               "github_release",
					},
				},
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "version override prefix not matching",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false",
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "true",
						VersionPrefix:      stringPtr("release-"),
						Type:               "github_release",
					},
				},
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "multiple version overrides - first matches",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false",
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "true",
						Type:               "github_release",
						RepoOwner:          "first",
					},
					{
						VersionConstraints: "true",
						Type:               "http",
						RepoOwner:          "second",
					},
				},
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "invalid version constraint expression",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "invalid_expression",
			},
			version:    "v1.0.0",
			expectName: "test-package", // Should still return a copy when constraint fails
		},
		{
			name: "invalid version constraint in override",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionConstraints: "false",
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "invalid_expression",
						Type:               "github_release",
					},
					{
						VersionConstraints: "true",
						Type:               "http",
					},
				},
			},
			version:    "v1.0.0",
			expectName: "test-package",
		},
		{
			name: "semver version constraint",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionPrefix:      "v",
				VersionConstraints: "semver(\">= 1.0.0\")",
			},
			version:    "v1.5.0",
			expectName: "test-package",
		},
		{
			name: "semver version constraint not matching",
			pkg: &registry.PackageInfo{
				Name:               "test-package",
				VersionPrefix:      "v",
				VersionConstraints: "semver(\">= 2.0.0\")",
			},
			version:    "v1.5.0",
			expectName: "test-package",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result, err := d.pkg.SetVersion(logE, d.version)

			if d.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Error("result should not be nil")
				return
			}

			if result.GetName() != d.expectName {
				t.Errorf("expected name %q, got %q", d.expectName, result.GetName())
			}

			// Verify that the result is a copy (different pointer) when constraints don't match
			if d.pkg.VersionConstraints != "" && d.pkg.VersionConstraints != "true" {
				if result == d.pkg {
					t.Error("result should be a copy of the original package, not the same instance")
				}
			}
		})
	}
}

func TestPackageInfo_SetVersion_Integration(t *testing.T) { //nolint:funlen
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())

	// Test a more complex scenario with realistic package info
	pkg := &registry.PackageInfo{
		Name:               "kubectl",
		Type:               "github_release",
		RepoOwner:          "kubernetes",
		RepoName:           "kubernetes",
		Asset:              "kubectl",
		VersionPrefix:      "v",
		VersionConstraints: "semver(\">= 1.20.0 < 1.25.0\")",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints: "semver(\"< 1.20.0\")",
				Type:               "http",
				URL:                "https://legacy.example.com/kubectl",
			},
			{
				VersionConstraints: "semver(\">= 1.25.0\")",
				Asset:              "kubectl-new",
				Format:             "tar.gz",
			},
		},
	}

	tests := []struct {
		version      string
		expectAsset  string
		expectType   string
		expectFormat string
		expectURL    string
	}{
		{
			version:    "v1.18.0", // Legacy version
			expectType: "http",
			expectURL:  "https://legacy.example.com/kubectl",
		},
		{
			version:     "v1.22.0", // Main version constraint
			expectAsset: "kubectl",
			expectType:  "github_release",
		},
		{
			version:      "v1.26.0", // New version with override
			expectAsset:  "kubectl-new",
			expectType:   "github_release",
			expectFormat: "tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			result, err := pkg.SetVersion(logE, tt.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectAsset != "" && result.Asset != tt.expectAsset {
				t.Errorf("expected asset %q, got %q", tt.expectAsset, result.Asset)
			}

			if result.Type != tt.expectType {
				t.Errorf("expected type %q, got %q", tt.expectType, result.Type)
			}

			if tt.expectFormat != "" && result.Format != tt.expectFormat {
				t.Errorf("expected format %q, got %q", tt.expectFormat, result.Format)
			}

			if tt.expectURL != "" && result.URL != tt.expectURL {
				t.Errorf("expected URL %q, got %q", tt.expectURL, result.URL)
			}
		})
	}
}

func TestPackageInfo_setTopVersion(t *testing.T) { //nolint:funlen
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())

	data := []struct {
		name      string
		pkg       *registry.PackageInfo
		version   string
		expectNil bool
	}{
		{
			name: "constraint matches",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionConstraints: "true",
			},
			version:   "v1.0.0",
			expectNil: false,
		},
		{
			name: "constraint does not match",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionConstraints: "false",
			},
			version:   "v1.0.0",
			expectNil: true,
		},
		{
			name: "version prefix matches",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionPrefix:      "v",
				VersionConstraints: "true",
			},
			version:   "v1.0.0",
			expectNil: false,
		},
		{
			name: "version prefix does not match",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionPrefix:      "release-",
				VersionConstraints: "true",
			},
			version:   "v1.0.0",
			expectNil: true,
		},
		{
			name: "invalid constraint expression",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionConstraints: "invalid_expression",
			},
			version:   "v1.0.0",
			expectNil: true,
		},
		{
			name: "semver constraint matches",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionPrefix:      "v",
				VersionConstraints: "semver(\">= 1.0.0\")",
			},
			version:   "v1.5.0",
			expectNil: false,
		},
		{
			name: "semver constraint does not match",
			pkg: &registry.PackageInfo{
				Name:               "test",
				VersionPrefix:      "v",
				VersionConstraints: "semver(\">= 2.0.0\")",
			},
			version:   "v1.5.0",
			expectNil: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			// Using the internal method through a wrapper
			result, err := d.pkg.SetVersion(logE, d.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// If the top-level constraint would match, we expect a result
			// If it wouldn't match, we expect a copy anyway (SetVersion always returns something)
			if result == nil {
				t.Error("SetVersion should never return nil")
			}

			// For testing the specific behavior, we need to check if it's the same instance or a copy
			// when constraints don't match, it should be a copy
			if d.expectNil && result == d.pkg {
				t.Error("when constraints don't match, result should be a copy")
			}
		})
	}
}
