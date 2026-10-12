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
	}
	for _, s := range words {
		if strings.Contains(asset, s) && !strings.Contains(pkgName, s) {
			return true
		}
	}
	return excludedExt(asset)
}

// excludedExt reports whether an asset is a file aqua can't install from: a package for
// a system package manager, or an SBOM published beside a binary.
func excludedExt(asset string) bool {
	exts := []string{
		".deb",
		".rpm",
		".msi",
		// SBOMs published beside each binary. regclient/regclient's v0.11.6 Windows
		// assets include regctl-windows-amd64.cyclonedx.json, which sorts before the
		// .exe and was taken for the binary.
		".cyclonedx.json",
		".cyclonedx.xml",
		".cdx.json",
		".cdx.xml",
		".bom.json",
		".bom.xml",
		".spdx.json",
		".spdx.xml",
		".spdx.yaml",
		".spdx.yml",
		".spdx.rdf",
		".spdx",
		".sbom.json",
		".sbom.xml",
		".sbom",
		".syft.json",
	}
	for _, ext := range exts {
		if strings.HasSuffix(asset, ext) {
			return true
		}
	}
	return false
}
