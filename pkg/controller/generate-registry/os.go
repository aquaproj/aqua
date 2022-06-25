package genrgst

import (
	"strings"
)

func (ctrl *Controller) setOS(assetName, lowAssetName string, assetInfo *AssetInfo) { //nolint:funlen
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
			break
		}
	}
	if assetInfo.OS == "" && strings.Contains(lowAssetName, ".exe") {
		assetInfo.OS = "windows"
	}
}
