package asset

import (
	"strings"

	"github.com/mholt/archiver/v3"
)

const formatRaw string = "raw"

// mholt/archiver/v3 not support but aqua support
func aquaSupportFormat(assetName string) string {
	if strings.HasSuffix(assetName, ".dmg") {
		return "dmg"
	}
	return formatRaw
}

func GetFormat(assetName string) string { //nolint:funlen,cyclop
	a, err := archiver.ByExtension(assetName)
	if err != nil {
		return aquaSupportFormat(assetName)
	}
	switch a.(type) {
	case *archiver.Rar:
		return "rar"
	case *archiver.Tar:
		return "tar"
	case *archiver.TarBrotli:
		if strings.HasSuffix(assetName, ".tbr") {
			return "tbr"
		}
		return "tar.br"
	case *archiver.TarBz2:
		if strings.HasSuffix(assetName, ".tbz2") {
			return "btz2"
		}
		return "tar.bz2"
	case *archiver.TarGz:
		if strings.HasSuffix(assetName, ".tgz") {
			return "tgz"
		}
		return "tar.gz"
	case *archiver.TarLz4:
		if strings.HasSuffix(assetName, ".tlz4") {
			return "tlz4"
		}
		return "tar.lz4"
	case *archiver.TarSz:
		if strings.HasSuffix(assetName, ".tsz") {
			return "tsz"
		}
		return "tar.sz"
	case *archiver.TarXz:
		if strings.HasSuffix(assetName, ".txz") {
			return "txz"
		}
		return "tar.xz"
	case *archiver.TarZstd:
		return "tar.zsd"
	case *archiver.Zip:
		return "zip"
	case *archiver.Gz:
		return "gz"
	case *archiver.Bz2:
		return "bz2"
	case *archiver.Lz4:
		return "lz4"
	case *archiver.Snappy:
		return "sz"
	case *archiver.Xz:
		return "xz"
	case *archiver.Zstd:
		return "zst"
	default:
		return aquaSupportFormat(assetName)
	}
}
