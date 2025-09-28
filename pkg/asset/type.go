// Package asset provides functionality for analyzing and parsing asset information
// from GitHub releases and other sources. It handles OS/architecture detection,
// file format identification, and template generation for package downloads.
package asset

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
