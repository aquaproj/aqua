package vacuum

import (
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfigPackageFromKey(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		key         string
		expected    *ConfigPackage
		expectError bool
	}{
		{
			name: "Valid key",
			key:  "github_release,test/pkg@v1.0.0",
			expected: &ConfigPackage{
				Type:    "github_release",
				Name:    "test/pkg",
				Version: "v1.0.0",
			},
			expectError: false,
		},
		{
			name:        "Invalid key format",
			key:         "invalid_key_format",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Empty key",
			key:         "",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := generateConfigPackageFromKey([]byte(tc.key))
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// GetPackageLastUsed retrieves the last used time of a package. for testing purposes.
func (vc *Controller) GetPackageLastUsed(logE *logrus.Entry, pkg *config.Package) *time.Time {
	var lastUsedTime time.Time
	vacuumpkg := vacuumConfigPackageFromConfigPackage(pkg)
	key := generateKey(vacuumpkg)
	pkgEntry, _ := vc.retrievePackageEntry(logE, key)
	if pkgEntry != nil {
		lastUsedTime = pkgEntry.LastUsageTime
	}
	return &lastUsedTime
}

// SetTimeStampPackage permit define a Timestamp for a package Manually. for testing purposes.
func (vc *Controller) SetTimestampPackages(logE *logrus.Entry, pkg []*config.Package, datetime time.Time) error {
	vacuumPkgs := make([]*ConfigPackage, 0, len(pkg))
	for _, p := range pkg {
		vacuumPkgs = append(vacuumPkgs, vacuumConfigPackageFromConfigPackage(p))
	}
	return vc.storePackageInternal(logE, vacuumPkgs, datetime)
}
