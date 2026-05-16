package asset

import (
	"strings"
)

// SetOS analyzes an asset name to detect and set operating system information.
// It matches common OS patterns in asset names, handles file extensions like .exe
// and .dmg for OS detection, and generates templates for cross-platform downloads.
// The function also manages OS name mappings and scoring for asset selection.
func SetOS(assetName, lowAssetName string, assetInfo *AssetInfo) { //nolint:funlen,cyclop
	if strings.Contains(lowAssetName, ".exe.") || strings.HasSuffix(lowAssetName, ".exe") {
		assetInfo.OS = osWindows
	} else if strings.HasSuffix(lowAssetName, ".dmg") || strings.HasSuffix(lowAssetName, ".pkg") {
		// other formats take precedence over DMG because DMG requires external commands.
		assetInfo.Score = -1
		assetInfo.OS = osDarwin
	}

	osList := []*OS{
		{
			Name: "apple-darwin",
			OS:   osDarwin,
		},
		{
			Name: "unknown-linux-gnu",
			OS:   osLinux,
		},
		{
			Name: "unknown-linux-musl",
			OS:   osLinux,
		},
		{
			Name: "unknown-linux",
			OS:   osLinux,
		},
		{
			Name: "linux-gnu",
			OS:   osLinux,
		},
		{
			Name: "pc-windows-msvc",
			OS:   osWindows,
		},
		{
			Name: "pc-windows-gnu",
			OS:   osWindows,
		},
		{
			Name: "pc-windows",
			OS:   osWindows,
		},
		{
			Name: osDarwin,
			OS:   osDarwin,
		},
		{
			Name: osLinux,
			OS:   osLinux,
		},
		{
			Name: osWindows,
			OS:   osWindows,
		},
		{
			Name: "apple",
			OS:   osDarwin,
		},
		{
			Name: "macosx",
			OS:   osDarwin,
		},
		{
			Name: "osx",
			OS:   osDarwin,
		},
		{
			Name: "macos",
			OS:   osDarwin,
		},
		{
			Name: "mac",
			OS:   osDarwin,
		},
		{
			Name: "win64",
			OS:   osWindows,
		},
		{
			Name: "win32",
			OS:   osWindows,
		},
		{
			Name: "win",
			OS:   osWindows,
		},
	}

	for _, o := range osList {
		if assetInfo.OS != "" && assetInfo.OS != o.OS {
			continue
		}
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 {
			assetInfo.OS = o.OS
			osName := assetName[idx : idx+len(o.Name)]
			if osName != o.OS {
				if assetInfo.Replacements == nil {
					assetInfo.Replacements = map[string]string{}
				}
				assetInfo.Replacements[o.OS] = osName
			}
			assetInfo.Template = strings.Replace(assetInfo.Template, osName, "{{.OS}}", 1)
			if osName == "unknown-linux-gnu" || osName == "pc-windows-gnu" {
				// "unknown-linux-musl" and "pc-windows-msvc" take precedence over "unknown-linux-gnu" and "pc-windows-gnu".
				assetInfo.Score = -1
			}
			return
		}
	}
}
