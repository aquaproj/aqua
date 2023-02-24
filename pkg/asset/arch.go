package asset

import (
	"strings"
)

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
			break
		}
	}
	if assetInfo.Arch == "" {
		assetInfo.Arch = "amd64"
		assetInfo.Score = -2 //nolint:gomnd
	}
}
