package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/google/go-cmp/cmp"
)

func Test_SetOS(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		assetName    string
		lowAssetName string
		assetInfo    *asset.AssetInfo
		exp          *asset.AssetInfo
	}{
		{
			name:         "no os",
			assetName:    "FOO.tar.gz",
			lowAssetName: "foo.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "FOO.tar.gz",
			},
			exp: &asset.AssetInfo{
				Template: "FOO.tar.gz",
			},
		},
		{
			name:         ".exe",
			assetName:    "foo.exe.gz",
			lowAssetName: "foo.exe.gz",
			assetInfo: &asset.AssetInfo{
				Template: "FOO.exe.gz",
			},
			exp: &asset.AssetInfo{
				OS:       "windows",
				Template: "FOO.exe.gz",
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
				OS:       "linux",
				Template: "FOO_{{.OS}}_AMD64.tar.gz",
				Replacements: map[string]string{
					"linux": "LINUX",
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
					"windows": "WINDOWS",
				},
			},
			exp: &asset.AssetInfo{
				OS:       "linux",
				Template: "FOO_{{.OS}}_AMD64.tar.gz",
				Replacements: map[string]string{
					"linux":   "LINUX",
					"windows": "WINDOWS",
				},
			},
		},
		{
			name:         "unknown-linux-gnu",
			assetName:    "silicon-v0.1.0-x86_64-unknown-linux-gnu.tar.gz",
			lowAssetName: "silicon-v0.1.0-x86_64-unknown-linux-gnu.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "silicon-{{.Version}}-x86_64-unknown-linux-gnu.tar.gz",
			},
			exp: &asset.AssetInfo{
				OS:       "linux",
				Template: "silicon-{{.Version}}-x86_64-{{.OS}}.tar.gz",
				Score:    -1,
				Replacements: map[string]string{
					"linux": "unknown-linux-gnu",
				},
			},
		},
		{
			name:         "pc-windows-gnu",
			assetName:    "silicon-v0.1.0-x86_64-pc-windows-gnu.tar.gz",
			lowAssetName: "silicon-v0.1.0-x86_64-pc-windows-gnu.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: "silicon-{{.Version}}-x86_64-pc-windows-gnu.tar.gz",
			},
			exp: &asset.AssetInfo{
				OS:       "windows",
				Template: "silicon-{{.Version}}-x86_64-{{.OS}}.tar.gz",
				Score:    -1,
				Replacements: map[string]string{
					"windows": "pc-windows-gnu",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			asset.SetOS(d.assetName, d.lowAssetName, d.assetInfo)
			if diff := cmp.Diff(d.assetInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
