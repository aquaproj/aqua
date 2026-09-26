package asset

import "strings"

// KnowsSpelling reports whether the parser reads spelling as that platform without being
// told.
//
// A registry's replacements say how a release writes a platform: darwin as osx, amd64 as
// x86_64. Most of what they say the parser works out for itself from the asset names, and
// a definition repeating it adds nothing to what is generated -- it is a copy of
// something that can be looked at, and one more thing to be wrong later.
//
// What it can't work out is a spelling nobody has taught it. luau-lang/luau calls its
// Linux build luau-ubuntu.zip, and there the replacement is the only way the asset
// belongs to an operating system at all. Asking this is how a definition can be reduced
// to that: what is left is what a release can't be read for.
//
// The platform is a GOOS or a GOARCH, and a spelling that is the platform's own name is
// known by definition.
func KnowsSpelling(platform, spelling string) bool {
	if platform == "" || spelling == "" {
		return false
	}
	// The parser matches asset names lowercased, so a spelling differing only in case is
	// the same spelling.
	name := strings.ToLower(spelling)
	if name == platform {
		return true
	}
	for _, o := range knownOSs() {
		if o.OS == platform && o.Name == name {
			return true
		}
	}
	for _, a := range knownArchs() {
		if a.Arch == platform && a.Name == name {
			return true
		}
	}
	return false
}
