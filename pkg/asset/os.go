package asset

import (
	"strings"
)

func SetOS(assetName, lowAssetName string, assetInfo *AssetInfo) { //nolint:funlen,cyclop
	osList := []*OS{
		{
			Name: "apple-darwin",
			OS:   "darwin",
		},
		{
			Name: "unknown-linux-gnu",
			OS:   "linux",
		},
		{
			Name: "unknown-linux-musl",
			OS:   "linux",
		},
		{
			Name: "pc-windows-msvc",
			OS:   "windows",
		},
		{
			Name: "pc-windows-gnu",
			OS:   "windows",
		},
		{
			Name: "darwin",
			OS:   "darwin",
		},
		{
			Name: "linux",
			OS:   "linux",
		},
		{
			Name: "windows",
			OS:   "windows",
		},
		{
			Name: "apple",
			OS:   "darwin",
		},
		{
			Name: "osx",
			OS:   "darwin",
		},
		{
			Name: "macos",
			OS:   "darwin",
		},
		{
			Name: "mac",
			OS:   "darwin",
		},
		{
			Name: "win64",
			OS:   "windows",
		},
		{
			Name: "win",
			OS:   "windows",
		},
	}
	for _, o := range osList {
		if idx := strings.Index(lowAssetName, o.Name); idx != -1 {
			osName := assetName[idx : idx+len(o.Name)]
			assetInfo.OS = o.OS
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
			break
		}
	}
	if assetInfo.OS == "" && strings.Contains(lowAssetName, ".exe") {
		assetInfo.OS = "windows"
	}
	if assetInfo.OS == "" && strings.HasSuffix(lowAssetName, ".dmg") {
		// other formats take precedence over DMG because DMG requires external commands.
		assetInfo.Score = -1
		assetInfo.OS = "darwin"
	}
}
