// Package asset provides functionality for analyzing and parsing asset information
// from GitHub releases and other sources. It handles OS/architecture detection,
// file format identification, and template generation for package downloads.
package asset

import "strings"

// OS represents an operating system mapping for asset parsing.
// Name is the string found in asset names, OS is the normalized OS name.
type OS struct {
	Name string
	OS   string
}

// Arch represents an architecture mapping for asset parsing.
// Name is the string found in asset names, Arch is the normalized architecture name.
type Arch struct {
	Name string
	Arch string
}

// AssetInfo contains parsed information about a release asset.
// It includes template generation data, platform detection results,
// format information, and scoring for asset selection.
type AssetInfo struct { //nolint:revive
	Template           string
	OS                 string
	Arch               string
	DarwinAll          bool
	Format             string
	Replacements       map[string]string
	Score              int
	CompleteWindowsExt *bool
}

// Spellings are names for a platform beyond the ones the parser knows.
//
// A release may write a platform in a way nothing here recognises —
// luau-lang/luau calls its Linux build luau-ubuntu.zip — and then the asset is read
// as belonging to no platform at all, so the package is generated without it. A
// registry that already says what the spelling means can say so here.
//
// The keys are the os or arch the spelling stands for; the values are how the
// release writes them.
type Spellings map[string]string

// osList returns the extra names for an operating system, to be tried before the
// ones the parser knows: a spelling given by name is more specific than one it
// guessed at.
func (s Spellings) osList() []*OS {
	out := make([]*OS, 0, len(s))
	for _, goos := range []string{osLinux, osDarwin, osWindows} {
		if name, ok := s[goos]; ok && name != "" && name != goos {
			out = append(out, &OS{Name: strings.ToLower(name), OS: goos})
		}
	}
	return out
}

// archList returns the extra names for an architecture.
func (s Spellings) archList() []*Arch {
	out := make([]*Arch, 0, len(s))
	for _, goarch := range []string{archAmd64, archArm64} {
		if name, ok := s[goarch]; ok && name != "" && name != goarch {
			out = append(out, &Arch{Name: strings.ToLower(name), Arch: goarch})
		}
	}
	return out
}
