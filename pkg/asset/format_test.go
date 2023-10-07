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
			name:      "tar.gz",
			assetName: "tfcmt_linux_amd64.tar.gz",
			exp:       "tfcmt_linux_amd64",
			format:    "tar.gz",
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
			format:    "raw",
		},
		{
			name:      "js",
			assetName: "tfcmt.js",
			exp:       "tfcmt.js",
			format:    "raw",
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
		d := d
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
