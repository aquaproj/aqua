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
			assetName:    assetFOOTarGz,
			lowAssetName: "foo.tar.gz",
			assetInfo: &asset.AssetInfo{
				Template: assetFOOTarGz,
			},
			exp: &asset.AssetInfo{
				Template: assetFOOTarGz,
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
				OS:       osWindows,
				Template: "FOO.exe.gz",
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
				OS:       osLinux,
				Template: "FOO_{{.OS}}_AMD64.tar.gz",
				Replacements: map[string]string{
					osLinux: osLinuxUpper,
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
					osWindows: "WINDOWS",
				},
			},
			exp: &asset.AssetInfo{
				OS:       osLinux,
				Template: "FOO_{{.OS}}_AMD64.tar.gz",
				Replacements: map[string]string{
					osLinux:   osLinuxUpper,
					osWindows: "WINDOWS",
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
				OS:       osLinux,
				Template: "silicon-{{.Version}}-x86_64-{{.OS}}.tar.gz",
				Score:    -1,
				Replacements: map[string]string{
					osLinux: "unknown-linux-gnu",
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
				OS:       osWindows,
				Template: "silicon-{{.Version}}-x86_64-{{.OS}}.tar.gz",
				Score:    -1,
				Replacements: map[string]string{
					osWindows: "pc-windows-gnu",
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			asset.SetOS(d.assetName, d.lowAssetName, d.assetInfo)
			if diff := cmp.Diff(d.assetInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
