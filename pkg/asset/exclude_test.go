package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func Test_Exclude(t *testing.T) { //nolint:funlen
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
			name:      "cyclonedx sbom",
			pkgName:   "regclient/regclient/regctl",
			assetName: "regctl-windows-amd64.cyclonedx.json",
			exp:       true,
		},
		{
			name:      "spdx sbom",
			pkgName:   "regclient/regclient/regctl",
			assetName: "regctl-windows-amd64.spdx.json",
			exp:       true,
		},
		{
			name:      "bom.json sbom",
			pkgName:   "foo/foo",
			assetName: "foo-windows-amd64.bom.json",
			exp:       true,
		},
		{
			name:      "cdx.xml sbom",
			pkgName:   "foo/foo",
			assetName: "foo-linux-amd64.cdx.xml",
			exp:       true,
		},
		{
			name:      "syft sbom",
			pkgName:   "foo/foo",
			assetName: "foo_linux_amd64.syft.json",
			exp:       true,
		},
		{
			name:      "the binary beside an sbom",
			pkgName:   "regclient/regclient/regctl",
			assetName: "regctl-windows-amd64.exe",
		},
		{
			name:      "gnueabi",
			pkgName:   "ekzhang/bore",
			assetName: "bore-v0.5.1-arm-unknown-linux-gnueabi.tar.gz",
			exp:       true,
		},
		{
			name:      "arm32",
			pkgName:   "XTLS/Xray-core",
			assetName: "Xray-linux-arm32-v5.zip",
			exp:       true,
		},
		{
			name:      "armhf",
			pkgName:   "ahmetb/kubectx",
			assetName: "kubectx_v0.9.4_windows_armhf.zip",
			exp:       true,
		},
		{
			name:      "arm64 is kept",
			pkgName:   "XTLS/Xray-core",
			assetName: "Xray-linux-arm64-v8a.zip",
		},
		{
			name:      "32bit",
			pkgName:   "suzuki-shunsuke/tfcmt",
			assetName: "tfcmt-linux-32bit.tar.gz",
			exp:       true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := asset.Exclude(d.pkgName, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
