package asset

import (
	"strings"
)

// SetArch analyzes an asset name to detect and set architecture information.
// It matches common architecture patterns in asset names and generates templates
// for cross-platform package downloads. The function also handles architecture
// name mappings and scoring for asset selection.
func SetArch(assetName, lowAssetName string, assetInfo *AssetInfo) {
	archList := []*Arch{
		{
			Name: "amd64",
			Arch: "amd64",
		},
		{
			Name: "arm64",
			Arch: "arm64",
		},
		{
			Name: "x86_64",
			Arch: "amd64",
		},
		{
			Name: "x64",
			Arch: "amd64",
		},
		{
			Name: "64bit",
			Arch: "amd64",
		},
		{
			Name: "64-bit",
			Arch: "amd64",
		},
		{
			Name: "aarch64",
			Arch: "arm64",
		},
		{
			Name: "arm",
			Arch: "arm64",
		},
	}
	for _, o := range archList {
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 {
			archName := assetName[idx : idx+len(o.Name)]
			assetInfo.Arch = o.Arch
			if archName != o.Arch {
				if assetInfo.Replacements == nil {
					assetInfo.Replacements = map[string]string{}
				}
				assetInfo.Replacements[o.Arch] = archName
			}
			assetInfo.Template = strings.Replace(assetInfo.Template, archName, "{{.Arch}}", 1)
			if o.Arch == "arm64" && o.Name == "arm" {
				assetInfo.Score -= 1
			}
			break
		}
	}
	if assetInfo.Arch == "" {
		assetInfo.Arch = "amd64"
		assetInfo.Score = -2 //nolint:mnd
	}
}
