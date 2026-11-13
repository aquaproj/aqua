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

	names := archNames(archList)
	for _, o := range archList {
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 && !shadowed(lowAssetName, idx, o.Name, names) {
			archName := assetName[idx : idx+len(o.Name)]
			assetInfo.Arch = o.Arch
			if archName != o.Arch {
				if assetInfo.Replacements == nil {
					assetInfo.Replacements = map[string]string{}
				}
				assetInfo.Replacements[o.Arch] = archName
			}
			assetInfo.Template = strings.Replace(assetInfo.Template, archName, "{{.Arch}}", 1)
			if o.Arch == archArm64 && o.Name == archArm {
				assetInfo.Score -= 1
			}
			break
		}
	}
	if assetInfo.Arch == "" {
		setNoArch(lowAssetName, assetInfo)
	}
}

// setNoArch sets an asset whose name says no architecture to amd64, which is what such a
// build usually is, and scores it below one whose name says so.
func setNoArch(lowAssetName string, assetInfo *AssetInfo) {
	assetInfo.Arch = archAmd64
	assetInfo.NoArch = true
	if !win64(lowAssetName, assetInfo) {
		assetInfo.Score = -2 //nolint:mnd
	}
}

// win64 says whether a Windows asset is named win64, which names the architecture as well
// as the OS: it is the x86-64 build, beside which win32 is the 32-bit one. Scored like a
// name without an architecture, the two tied, and protoc and scrcpy got their 32-bit builds
// for windows/amd64.
func win64(lowAssetName string, assetInfo *AssetInfo) bool {
	return assetInfo.OS == osWindows && strings.Contains(lowAssetName, "win64")
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
		{Name: archArm, Arch: archArm64},
	}
}

// shadowed says whether the name found at idx is only part of a longer name found around
// it.
//
// The names are tried in their order, and a shorter one can be contained in a longer one:
// XTLS/Xray-core spells amd64 as 64 and arm64 as arm64-v8a, and 64 found inside
// Xray-linux-arm64-v8a.zip read the asset as amd64. The longer name, wherever it sits in
// the order, is the one the asset says.
func shadowed(lowAssetName string, idx int, name string, names []string) bool {
	for _, other := range names {
		if len(other) <= len(name) || !strings.Contains(other, name) {
			continue
		}
		for start := 0; ; {
			i := strings.Index(lowAssetName[start:], other)
			if i == -1 {
				break
			}
			i += start
			if i <= idx && idx+len(name) <= i+len(other) {
				return true
			}
			start = i + 1
		}
	}
	return false
}

func archNames(list []*Arch) []string {
	names := make([]string, len(list))
	for i, o := range list {
		names[i] = o.Name
	}
	return names
}
