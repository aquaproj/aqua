package asset

import (
	"context"
	"strings"

	"github.com/mholt/archives"
)

const (
	formatRaw string = "raw"
	formatZip string = "zip"
)

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

		formatZip,
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
		if s, ok := strings.CutSuffix(assetName, "."+format); ok {
			return s, format
		}
	}
	if format := archiveFormat(assetName); format != "" {
		return strings.TrimSuffix(assetName, "."+format), format
	}
	return assetName, formatRaw
}

// archiveFormat returns the format the unarchiver opens an asset as, read off its name, or
// an empty string when it opens none.
//
// The list above is how a name is read first: it has the short forms such as tgz, and dmg
// and pkg, which mholt/archives doesn't, and it decides what AssetWithoutExt drops. The
// unarchiver is mholt/archives, though, so a format it opens is one aqua can install, and
// asking it keeps the formats it adds, such as 7z, from being taken for raw files.
//
// Identify matches a name that merely contains an extension: foo.tar.gz.sha256 is a
// .tar.gz to it. Only a name that ends in the format's extension is the format here.
//
// Identify is given no stream, so it reads the name and nothing else, and the context it
// takes has nothing to cancel.
func archiveFormat(assetName string) string {
	format, _, err := archives.Identify(context.Background(), assetName, nil)
	if err != nil {
		return ""
	}
	ext := format.Extension()
	if ext == "" || !strings.HasSuffix(assetName, ext) {
		return ""
	}
	return strings.TrimPrefix(ext, ".")
}

// getFormat extracts the file format from an asset name.
// It uses RemoveExtFromAsset to determine the format type.
func getFormat(assetName string) string {
	_, format := RemoveExtFromAsset(assetName)
	return format
}
