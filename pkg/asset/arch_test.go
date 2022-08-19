package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/asset"
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
			assetName:    "FOO.tar.gz",
			lowAssetName: "foo.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "FOO.tar.gz",
			},
			exp: &asset.AssetInfo{
				Template: "FOO.tar.gz",
				Arch:     "amd64",
				Score:    -2,
			},
		},
		{
			name:         "FOO_LINUX_AMD64.tar.gz",
			assetName:    "FOO_LINUX_AMD64.tar.gz",
			lowAssetName: "foo_linux_amd64.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "FOO_LINUX_AMD64.tar.gz",
			},
			exp: &asset.AssetInfo{
				Arch:     "amd64",
				Template: "FOO_LINUX_{{.Arch}}.tar.gz",
				Replacements: map[string]string{
					"amd64": "AMD64",
				},
			},
		},
		{
			name:         "FOO_LINUX_AMD64.tar.gz non nil replacements",
			assetName:    "FOO_LINUX_AMD64.tar.gz",
			lowAssetName: "foo_linux_amd64.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "FOO_LINUX_AMD64.tar.gz",
				Replacements: map[string]string{
					"linux": "LINUX",
				},
			},
			exp: &asset.AssetInfo{
				Arch:     "amd64",
				Template: "FOO_LINUX_{{.Arch}}.tar.gz",
				Replacements: map[string]string{
					"linux": "LINUX",
					"amd64": "AMD64",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			asset.SetArch(d.assetName, d.lowAssetName, d.assetInfo)
			if diff := cmp.Diff(d.assetInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
