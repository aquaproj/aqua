package unarchive

import (
	"path/filepath"
	"strings"
)

// SingleFileName returns the name the extraction writes when the asset is one
// compressed file rather than an archive of entries, and an empty string when it is an
// archive.
//
// A compressed file holds no name of its own: what is written is the asset's name
// without the extension that said it was compressed. So what a package installs from is
// known before anything is downloaded, the same way it is for an asset that isn't
// compressed at all, and a definition needs no files[].src to say it.
func SingleFileName(format, assetName string) string {
	filename := filepath.Base(assetName)
	if !singleFile(format, filename) {
		return ""
	}
	return decompressedName(filename)
}

// decompressedName is the name a decompressed single file is written as. It is here
// rather than at the one place that writes it, so that what says where a package
// installs from and what puts it there cannot disagree.
func decompressedName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// singleFile says the asset is one compressed file.
//
// The format is asked first, because that is what the definition declares and what
// decides how the asset is opened; the extension answers for a definition that declares
// nothing. What the two can't settle is a tar: "foo.tar.gz" is gzip of an archive, so
// the name left after the compression extension tells them apart.
func singleFile(format, filename string) bool {
	ext := strings.TrimPrefix(strings.ToLower(format), ".")
	if ext == "" {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	}
	switch ext {
	case "gz", "bz2", "xz", "zst", "lz4", "br", "sz", "z":
	default:
		return false
	}
	return !strings.EqualFold(filepath.Ext(decompressedName(filename)), ".tar")
}
