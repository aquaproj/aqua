package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/asset"
)

func TestGetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		assetName string
		exp       string
	}{
		{
			name:      "tar.gz",
			assetName: "tfcmt_linux_amd64.tar.gz",
			exp:       "tar.gz",
		},
		{
			name:      "tgz",
			assetName: "tfcmt_linux_amd64.tgz",
			exp:       "tgz",
		},
		{
			name:      "exe",
			assetName: "tfcmt_windows_amd64.exe",
			exp:       "raw",
		},
		{
			name:      "js",
			assetName: "tfcmt.js",
			exp:       "raw",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			format := asset.GetFormat(d.assetName)
			if format != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, format)
			}
		})
	}
}
