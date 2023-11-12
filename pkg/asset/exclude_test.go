package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
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
		},
		{
			name:      "32bit",
			pkgName:   "suzuki-shunsuke/tfcmt",
			assetName: "tfcmt-linux-32bit.tar.gz",
			exp:       true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := asset.Exclude(d.pkgName, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
