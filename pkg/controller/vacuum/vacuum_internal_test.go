package vacuum

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestHandleAsyncStorePackage_NilPackage(t *testing.T) {
	t.Parallel()
	logE := logrus.NewEntry(logrus.New())

	vacuumCtrl := New(context.Background(), nil, nil)

	// Test
	err := vacuumCtrl.handleAsyncStorePackage(logE, nil)

	// Assert
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}
	if diff := cmp.Diff("vacuumPkg is nil", err.Error()); diff != "" {
		t.Errorf("unexpected error message (-want +got):\n%s", diff)
	}
}

func TestEncodePackageEntry(t *testing.T) {
	t.Parallel()
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := encodePackageEntry(pkgEntry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data == nil {
		t.Fatalf("expected data but got nil")
	}

	var decodedEntry PackageEntry
	if err = json.Unmarshal(data, &decodedEntry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix()); diff != "" {
		t.Errorf("unexpected LastUsageTime (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(pkgEntry.Package.Name, decodedEntry.Package.Name); diff != "" {
		t.Errorf("unexpected Package.Name (-want +got):\n%s", diff)
	}
}

func TestDecodePackageEntry(t *testing.T) {
	t.Parallel()
	pkgEntry := &PackageEntry{
		LastUsageTime: time.Now(),
		Package:       &Package{Name: "test-package"},
	}

	data, err := json.Marshal(pkgEntry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decodedEntry, err := decodePackageEntry(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decodedEntry == nil {
		t.Fatalf("expected decodedEntry but got nil")
	}
	if diff := cmp.Diff(pkgEntry.LastUsageTime.Unix(), decodedEntry.LastUsageTime.Unix()); diff != "" {
		t.Errorf("unexpected LastUsageTime (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(pkgEntry.Package.Name, decodedEntry.Package.Name); diff != "" {
		t.Errorf("unexpected Package.Name (-want +got):\n%s", diff)
	}
}

func TestDecodePackageEntry_Error(t *testing.T) {
	t.Parallel()
	_, err := decodePackageEntry([]byte("invalid json"))
	if err == nil {
		t.Fatalf("expected an error but got nil")
	}
	if diff := cmp.Diff(true, strings.HasPrefix(err.Error(), "unmarshal package entry")); diff != "" {
		t.Errorf("unexpected error message (-want +got):\n%s", diff)
	}
}
