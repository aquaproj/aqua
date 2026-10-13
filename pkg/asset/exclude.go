package asset

import (
	"strings"
)

// Exclude determines whether an asset should be excluded from consideration.
// It filters out assets for unsupported architectures, platforms, and file types
// that are not compatible with aqua's installation process.
func Exclude(pkgName, assetName string) bool {
	asset := strings.ToLower(assetName)
	pkg := strings.ToLower(pkgName)
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
		"netbsd",
		"openbsd",
		"plan9",
		"ppc64",
		"readme",
		"riscv64",
		"s390x",
		"solaris",
		"wasm",
	}
	for _, s := range words {
		if strings.Contains(asset, s) && !strings.Contains(pkg, s) {
			return true
		}
	}
	if s := arm32Spelling(asset); s != "" && !strings.Contains(pkg, s) {
		return true
	}
	return excludedExt(asset)
}

// arm32Tokens are words of an asset name that say 32-bit ARM.
//
// The parser reads "arm" as arm64, so a 32-bit ARM build that isn't excluded becomes the
// arm64 asset when no arm64 build is found: XTLS/Xray-core's arm32-v5, ahmetb/kubectx's
// windows_armhf. "arm" itself isn't one of them, since some releases write their 64-bit
// builds that way.
var arm32Tokens = map[string]struct{}{ //nolint:gochecknoglobals
	"arm32": {}, "armel": {}, "armhf": {}, "armv5": {}, "armv5te": {}, "armv8l": {},
	"arm5": {}, "arm6": {}, "arm7": {},
}

// arm32Spelling returns the 32-bit ARM spelling an asset name has, or "".
//
// The name is read word by word, so that a spelling has to be a whole word: arm_64 and
// arm-64 are 64-bit, and neither carmel nor seabird is a spelling. A word ending in eabi
// is an ARM embedded ABI (gnueabi, musleabi, uclibceabi), and arm followed by v5, v6 or v7
// is a version written apart (arm-v7).
func arm32Spelling(asset string) string {
	words := strings.FieldsFunc(asset, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	for i, w := range words {
		if _, ok := arm32Tokens[w]; ok {
			return w
		}
		if strings.HasSuffix(w, "eabi") {
			return w
		}
		if w == archArm && i+1 < len(words) {
			switch words[i+1] {
			case "v5", "v6", "v7":
				return w + "-" + words[i+1]
			}
		}
	}
	return ""
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
