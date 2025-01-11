package vacuum

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAsyncStorePackage_NilPackage(t *testing.T) {
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())

	vacuumCtrl := New(context.Background(), nil, nil)

	// Test
	err := vacuumCtrl.handleAsyncStorePackage(logE, nil)

	// Assert
	require.Error(t, err)
	assert.Equal(t, "vacuumPkg is nil", err.Error())
}

func TestEncodePackageEntry(t *testing.T) {
	t.Parallel()
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := encodePackageEntry(pkgEntry)
	require.NoError(t, err)
	assert.NotNil(t, data)

	var decodedEntry PackageEntry
	err = json.Unmarshal(data, &decodedEntry)
	require.NoError(t, err)
	assert.Equal(t, pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix())
	assert.Equal(t, pkgEntry.Package.Name, decodedEntry.Package.Name)
}

func TestDecodePackageEntry(t *testing.T) {
	t.Parallel()
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := json.Marshal(pkgEntry)
	require.NoError(t, err)

	decodedEntry, err := decodePackageEntry(data)
	require.NoError(t, err)
	assert.NotNil(t, decodedEntry)
	assert.Equal(t, pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix())
	assert.Equal(t, pkgEntry.Package.Name, decodedEntry.Package.Name)
}

func TestDecodePackageEntry_Error(t *testing.T) {
	t.Parallel()
	_, err := decodePackageEntry([]byte("invalid json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal package entry")
}
