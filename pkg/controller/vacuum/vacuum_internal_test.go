package vacuum

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHandleAsyncStorePackage_NilPackage(t *testing.T) {
	logE := logrus.NewEntry(logrus.New())

	vacuumCtrl := New(nil, nil)

	// Test
	err := vacuumCtrl.handleAsyncStorePackage(logE, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "vacuumPkg is nil", err.Error())
}
func TestEncodePackageEntry(t *testing.T) {
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := encodePackageEntry(pkgEntry)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var decodedEntry PackageEntry
	err = json.Unmarshal(data, &decodedEntry)
	assert.NoError(t, err)
	assert.Equal(t, pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix())
	assert.Equal(t, pkgEntry.Package.Name, decodedEntry.Package.Name)
}

func TestDecodePackageEntry(t *testing.T) {
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := json.Marshal(pkgEntry)
	assert.NoError(t, err)

	decodedEntry, err := decodePackageEntry(data)
	assert.NoError(t, err)
	assert.NotNil(t, decodedEntry)
	assert.Equal(t, pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix())
	assert.Equal(t, pkgEntry.Package.Name, decodedEntry.Package.Name)
}

func TestDecodePackageEntry_Error(t *testing.T) {
	_, err := decodePackageEntry([]byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal package entry")
}
