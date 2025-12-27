package asset

import (
	"strings"
)

// Exclude determines whether an asset should be excluded from consideration.
// It filters out assets for unsupported architectures, platforms, and file types
// that are not compatible with aqua's installation process.
func Exclude(pkgName, assetName string) bool {
	asset := strings.ToLower(assetName)
	words := []string{
		"32-bit",
		"32bit",
		"386",
		"android",
		"armv6",
		"armv7",
		"changelog",
		"eabihf",
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
	exts := []string{
		".deb",
		".rpm",
		".msi",
	}
	for _, ext := range exts {
		if strings.HasSuffix(asset, ext) {
			return true
		}
	}
	return false
}
