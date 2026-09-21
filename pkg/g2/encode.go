// Package g2 reads aqua-registry-g2, the registry that serves statically resolved
// registry.json instead of templates.
//
// Each package has its own orphan branch holding one registry.json per version, so
// aqua fetches exactly the file it needs rather than a registry covering every
// package. That is what lets a lock file be built without downloading a registry
// whose other entries are irrelevant.
package g2

import (
	"fmt"
	"strings"
)

// BranchPrefix marks the branches holding a package's generated registry.json.
const BranchPrefix = "pkg_"

// VersionDir is the directory those files live in on the branch.
const VersionDir = "versions"

// FileName is the generated file a version directory holds.
const FileName = "registry.json"

// BranchName returns the branch holding the package's generated registry.json.
func BranchName(pkgName string) string {
	return BranchPrefix + EncodePackageName(pkgName)
}

// Path returns where the package's registry.json sits on its branch.
func Path(version string) string {
	return VersionDir + "/" + version + "/" + FileName
}

// EncodePackageName escapes a package name so it can be used as a git ref.
//
// Every character outside [A-Za-z0-9.-] becomes an underscore followed by two
// lowercase hex digits. The underscore is escaped too, which is what makes the
// mapping reversible where turning a slash into two underscores is not: package
// names contain underscores.
//
//	cli/cli                    -> cli_2fcli
//	ipinfo/cli/grepip          -> ipinfo_2fcli_2fgrepip
//	sr.ht/~charles/rq          -> sr.ht_2f_7echarles_2frq
//	sue445/plant_erd           -> sue445_2fplant_5ferd
//
// The result stays within [A-Za-z0-9._-], so it needs no further escaping in a URL.
// Percent-encoding would: a branch named with "%7E" has to be written "%257E" in a
// raw URL, because the server decodes the escape before looking the ref up.
//
// It also contains no slash, so it is a single ref segment. That removes the
// directory/file conflict which would otherwise stop "ipinfo/cli" and
// "ipinfo/cli/grepip" from existing as branches at the same time.
func EncodePackageName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, c := range []byte(name) {
		if isSafe(c) {
			b.WriteByte(c)
			continue
		}
		fmt.Fprintf(&b, "_%02x", c)
	}
	return b.String()
}

func isSafe(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z',
		c >= 'a' && c <= 'z',
		c >= '0' && c <= '9',
		c == '.', c == '-':
		return true
	}
	return false
}
