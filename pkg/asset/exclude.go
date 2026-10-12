package asset

import (
	"regexp"
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
		"changelog",
		// ARM's embedded ABIs, all 32-bit: gnueabi, gnueabihf, musleabi, uclibceabi.
		"eabi",
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
	if arm32Name.MatchString(asset) && !arm32Name.MatchString(pkg) {
		return true
	}
	return excludedExt(asset)
}

// arm32Name matches a 32-bit ARM spelling in an asset name.
//
// The parser reads "arm" as arm64, so a 32-bit ARM build that isn't excluded becomes the
// arm64 asset when no arm64 build is found: XTLS/Xray-core's arm32-v5, ahmetb/kubectx's
// windows_armhf. "arm" itself stays, since some releases write their 64-bit builds that
// way. It has to start a word, so that carmel isn't armel, and arm6 is left out because
// it is how arm64 begins.
var arm32Name = regexp.MustCompile(`(^|[^a-z])arm(32|el|hf|v[5-7]|v8l|[57]([^0-9]|$)|[-_]v?[5-7])`)

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
