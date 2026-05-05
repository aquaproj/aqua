package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/google/go-cmp/cmp"
)

func Test_SetArch(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		assetName    string
		lowAssetName string
		assetInfo    *asset.AssetInfo
		exp          *asset.AssetInfo
	}{
		{
			name:         "no arch",
			assetName:    assetFOOTarGz,
			lowAssetName: "foo.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: assetFOOTarGz,
			},
			exp: &asset.AssetInfo{
				Template: assetFOOTarGz,
				Arch:     archAmd64,
				Score:    -2,
			},
		},
		{
			name:         assetFOOLinuxAmd64TarGz,
			assetName:    assetFOOLinuxAmd64TarGz,
			lowAssetName: assetFooLinuxAmd64TarGz,
			assetInfo: &asset.AssetInfo{
				Template: assetFOOLinuxAmd64TarGz,
			},
			exp: &asset.AssetInfo{
				Arch:     archAmd64,
				Template: "FOO_LINUX_{{.Arch}}.tar.gz",
				Replacements: map[string]string{
					archAmd64: "AMD64",
				},
			},
		},
		{
			name:         "FOO_LINUX_AMD64.tar.gz non nil replacements",
			assetName:    assetFOOLinuxAmd64TarGz,
			lowAssetName: assetFooLinuxAmd64TarGz,
			assetInfo: &asset.AssetInfo{
				Template: assetFOOLinuxAmd64TarGz,
				Replacements: map[string]string{
					osLinux: osLinuxUpper,
				},
			},
			exp: &asset.AssetInfo{
				Arch:     archAmd64,
				Template: "FOO_LINUX_{{.Arch}}.tar.gz",
				Replacements: map[string]string{
					osLinux:   osLinuxUpper,
					archAmd64: "AMD64",
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			asset.SetArch(d.assetName, d.lowAssetName, d.assetInfo)
			if diff := cmp.Diff(d.assetInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
