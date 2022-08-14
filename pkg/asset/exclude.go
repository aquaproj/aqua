package asset

import (
	"strings"

	"github.com/aquaproj/aqua/pkg/util"
)

func Exclude(pkgName, assetName, version string) bool {
	allowedExts := map[string]struct{}{
		".exe": {},
		".sh":  {},
		".js":  {},
		".jar": {},
		".py":  {},
	}
	if format := GetFormat(assetName); format == formatRaw {
		ext := util.Ext(assetName, version)
		if len(ext) > 0 && len(ext) < 6 {
			if _, ok := allowedExts[ext]; !ok {
				return true
			}
		}
	}
	suffixes := []string{
		"sha256",
	}
	asset := strings.ToLower(assetName)
	for _, s := range suffixes {
		if strings.HasSuffix(asset, "."+s) {
			return true
		}
	}
	words := []string{
		"32-bit",
		"32bit",
		"386",
		"android",
		"armv6",
		"armv7",
		"changelog",
		"checksum",
		"freebsd",
		"i386",
		"license",
		"mips",
		"mips64",
		"mips64le",
		"mipsle",
		"netbsd",
		"netbsd",
		"openbsd",
		"plan9",
		"ppc64",
		"ppc64le",
		"readme",
		"riscv64",
		"s390x",
		"solaris",
		"wasm",
		"win32",
	}
	for _, s := range words {
		if strings.Contains(asset, s) && !strings.Contains(pkgName, s) {
			return true
		}
	}
	return false
}
