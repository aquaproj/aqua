package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func TestRemoveExtFromAsset(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		assetName string
		exp       string
		format    string
	}{
		{
			name:      formatTarGz,
			assetName: "tfcmt_linux_amd64.tar.gz",
			exp:       "tfcmt_linux_amd64",
			format:    formatTarGz,
		},
		{
			name:      "tgz",
			assetName: "tfcmt_linux_amd64.tgz",
			exp:       "tfcmt_linux_amd64",
			format:    "tgz",
		},
		{
			name:      "exe",
			assetName: "tfcmt_windows_amd64.exe",
			exp:       "tfcmt_windows_amd64.exe",
			format:    formatRaw,
		},
		{
			name:      "js",
			assetName: "tfcmt.js",
			exp:       "tfcmt.js",
			format:    formatRaw,
		},
		{
			name:      "dmg",
			assetName: "aws-vault-darwin-amd64.dmg",
			exp:       "aws-vault-darwin-amd64",
			format:    "dmg",
		},
		{
			name:      "pkg",
			assetName: "aws-vault-darwin-amd64.pkg",
			exp:       "aws-vault-darwin-amd64",
			format:    "pkg",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			assetWithoutExt, format := asset.RemoveExtFromAsset(d.assetName)
			if assetWithoutExt != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, assetWithoutExt)
			}
			if format != d.format {
				t.Fatalf("wanted %v, got %v", d.format, format)
			}
		})
	}
}

// TestRemoveExtFromAssetArchives covers the formats read off mholt/archives rather than
// aqua's own list.
func TestRemoveExtFromAssetArchives(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		assetName string
		exp       string
		format    string
	}{
		{
			name:      "7z",
			assetName: "ImageMagick-7.1.2-36-portable-Q16-HDRI-x64.7z",
			exp:       "ImageMagick-7.1.2-36-portable-Q16-HDRI-x64",
			format:    "7z",
		},
		{
			name:      "tar.lz",
			assetName: "foo_linux_amd64.tar.lz",
			exp:       "foo_linux_amd64",
			format:    "tar.lz",
		},
		{
			name:      "checksum of an archive",
			assetName: "foo_linux_amd64.tar.gz.sha256",
			exp:       "foo_linux_amd64.tar.gz.sha256",
			format:    formatRaw,
		},
		{
			name:      "signature of a 7z",
			assetName: "foo_windows_amd64.7z.sig",
			exp:       "foo_windows_amd64.7z.sig",
			format:    formatRaw,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			assetWithoutExt, format := asset.RemoveExtFromAsset(d.assetName)
			if assetWithoutExt != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, assetWithoutExt)
			}
			if format != d.format {
				t.Fatalf("wanted %v, got %v", d.format, format)
			}
		})
	}
}
