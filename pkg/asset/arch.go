package asset

import (
	"strings"
)

// SetArch analyzes an asset name to detect and set architecture information.
// It matches common architecture patterns in asset names and generates templates
// for cross-platform package downloads. The function also handles architecture
// name mappings and scoring for asset selection.
func SetArch(assetName, lowAssetName string, assetInfo *AssetInfo, extra ...Spellings) {
	archList := knownArchs()
	for _, s := range extra {
		archList = append(s.archList(), archList...)
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
			if o.Arch == archArm64 && o.Name == "arm" {
				assetInfo.Score -= 1
			}
			break
		}
	}
	if assetInfo.Arch == "" {
		assetInfo.Arch = archAmd64
		assetInfo.Score = -2 //nolint:mnd
	}
}

// knownArchs are the spellings of an architecture the parser recognises on its own.
//
// The order matters: a name that contains another has to come first, or the shorter one
// matches the longer one's asset.
func knownArchs() []*Arch {
	return []*Arch{
		{Name: archAmd64, Arch: archAmd64},
		{Name: archArm64, Arch: archArm64},
		{Name: archX86_64, Arch: archAmd64},
		{Name: "x86-64", Arch: archAmd64},
		{Name: "x64", Arch: archAmd64},
		{Name: "64bit", Arch: archAmd64},
		{Name: "64-bit", Arch: archAmd64},
		{Name: "aarch64", Arch: archArm64},
		{Name: "arm", Arch: archArm64},
	}
}
