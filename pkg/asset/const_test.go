package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func TestConstants(t *testing.T) {
	t.Parallel()

	// Test that constants are accessible and have expected values
	// This is mainly for coverage, but also ensures the constants don't change unexpectedly

	// Test by parsing an asset that should use raw format
	assetInfo := asset.ParseAssetName("tool.exe", "1.0.0")
	if assetInfo.Format != "raw" {
		t.Errorf("Expected raw format for exe file, got %s", assetInfo.Format)
	}

	// Test by parsing an asset with tar.gz format
	assetInfo2 := asset.ParseAssetName("tool-1.0.0.tar.gz", "1.0.0")
	if assetInfo2.Format != "tar.gz" {
		t.Errorf("Expected tar.gz format, got %s", assetInfo2.Format)
	}
}
