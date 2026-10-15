// Package g2 reads aqua-registry-g2, the registry that serves statically resolved
// registry.json instead of templates.
//
// Each package has its own directory holding one registry.json per version, so aqua
// fetches exactly the file it needs rather than a registry covering every package. That is what lets a lock file be built without downloading a registry
// whose other entries are irrelevant.
package g2

import (
	"fmt"
	"strconv"
	"strings"
)

// PackagesDir is the directory on the default branch holding every package.
const PackagesDir = "pkgs"

// VersionDir is the directory in a package's directory those files live in.
const VersionDir = "versions"

// FileName is the generated file a version directory holds.
//
// The name carries the schema's major version. aqua fetches this file by path and has
// no ref to pin, so a name that never changed would one day hand an older aqua a
// schema it cannot read, with nothing to fall back to. A new major is a new name:
// aqua asks for the newest it knows and falls back to older ones, which is what lets
// a schema change reach users gradually instead of all at once.
//
// Minor additions keep the name, because the same reader reads them. The full version
// is recorded inside the file.
const FileName = "registry-1.json"

// PackageDir returns the directory holding the package whose id this is, such as
// pkgs/69/1790772769.
//
// The directory is named after the package's id rather than after the package, so that a
// repository being renamed moves nothing: the name is resolved to an id through the table
// the registry publishes, and everything else addresses the id. The id's last two digits
// spread the packages over a hundred directories, since an id is the second it was minted.
func PackageDir(id string) string {
	return PackagesDir + "/" + Shard(id) + "/" + id
}

// shardWidth is how many of an id's trailing digits name its shard.
const shardWidth = 2

// Shard is the directory under PackagesDir the package whose id this is sits in.
func Shard(id string) string {
	if len(id) < shardWidth {
		return strings.Repeat("0", shardWidth-len(id)) + id
	}
	return id[len(id)-shardWidth:]
}

// Path returns where the package's registry.json sits in the package's directory.
func Path(version string) string {
	return VersionDir + "/" + EncodeVersion(version) + "/" + FileName
}

// EncodePackageName escapes a package name so that it is one path segment, which is how the
// cache keeps a package's files under its name.
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
// The result stays within [A-Za-z0-9._-] and contains no slash, so "ipinfo/cli" and
// "ipinfo/cli/grepip" are two directories side by side rather than one inside the other.
func EncodePackageName(name string) string {
	return encode(name)
}

// EncodeVersion escapes a version so that it is one path segment.
//
// A version is not always a plain "v1.2.3": a monorepo tags "kustomize/v5.8.1", and a
// package published from a JavaScript monorepo tags "@yarnpkg/cli/4.16.0". Written
// into a path as it is, the slash becomes a directory, so one version's directory is
// another's parent and the versions a package has can no longer be told from the first
// segments of their tags: every kustomize release is a directory called "kustomize".
//
//	v1.2.3             -> v1.2.3
//	kustomize/v5.8.1   -> kustomize_2fv5.8.1
//	@yarnpkg/cli/4.16.0 -> _40yarnpkg_2fcli_2f4.16.0
//
// It is the escaping a package name uses, with the same reversibility. A version that
// needs no escape is left as it is, which is nearly all of them.
func EncodeVersion(version string) string {
	return encode(version)
}

// DecodeVersion undoes EncodeVersion.
//
// It reports false for anything EncodeVersion couldn't have produced, which is what
// tells a version apart from a directory left by the layout that wrote a version's
// slashes as directories.
func DecodeVersion(encoded string) (string, bool) {
	return decode(encoded)
}

func encode(name string) string {
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

// DecodePackageName undoes EncodePackageName.
//
// It reports false for anything EncodePackageName couldn't have produced, such as a
// truncated escape or a character that would never have been escaped, rather than
// returning a name that was never encoded.
func DecodePackageName(encoded string) (string, bool) {
	return decode(encoded)
}

func decode(encoded string) (string, bool) {
	var b strings.Builder
	b.Grow(len(encoded))
	for i := 0; i < len(encoded); i++ {
		c := encoded[i]
		if c != '_' {
			if !isSafe(c) {
				return "", false
			}
			b.WriteByte(c)
			continue
		}
		if i+2 >= len(encoded) {
			return "", false
		}
		n, err := strconv.ParseUint(encoded[i+1:i+3], 16, 8)
		if err != nil {
			return "", false
		}
		decoded := byte(n)
		// Only what would have been escaped: an escape of a safe character is
		// something else's text, not a name this produced.
		if isSafe(decoded) {
			return "", false
		}
		b.WriteByte(decoded)
		i += 2
	}
	return b.String(), true
}
