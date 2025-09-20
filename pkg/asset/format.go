package asset

import (
	"strings"
)

const formatRaw string = "raw"

// RemoveExtFromAsset removes file extensions from asset names and returns the format.
// It recognizes various archive formats including tar variants, compression formats,
// and platform-specific formats. Returns the asset name without extension and the format.
func RemoveExtFromAsset(assetName string) (string, string) {
	formats := []string{
		"tar.br",
		"tar.bz2",
		"tar.gz",
		"tar.lz4",
		"tar.sz",
		"tar.xz",
		"tbr",
		"tbz",
		"tbz2",
		"tgz",
		"tlz4",
		"tsz",
		"txz",

		"tar.zst",

		"zip",
		"gz",
		"bz2",
		"lz4",
		"sz",
		"xz",
		"zst",

		"dmg",
		"pkg",

		"rar",
		"tar",
	}
	for _, format := range formats {
		if s := strings.TrimSuffix(assetName, "."+format); s != assetName {
			return s, format
		}
	}
	return assetName, "raw"
}

// getFormat extracts the file format from an asset name.
// It uses RemoveExtFromAsset to determine the format type.
func getFormat(assetName string) string {
	_, format := RemoveExtFromAsset(assetName)
	return format
}
