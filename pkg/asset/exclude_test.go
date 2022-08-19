package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/asset"
)

func Test_Exclude(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		pkgName   string
		assetName string
		version   string
		exp       bool
	}{
		{
			name:      "not exclude",
			pkgName:   "suzuki-shunsuke/tfcmt",
			assetName: "tfcmt_linux_amd64.tar.gz",
			version:   "v3.0.0",
		},
		{
			name:      "32bit",
			pkgName:   "suzuki-shunsuke/tfcmt",
			assetName: "tfcmt-linux-32bit.tar.gz",
			version:   "v3.0.0",
			exp:       true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := asset.Exclude(d.pkgName, d.assetName, d.version)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
