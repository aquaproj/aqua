package which

import (
	"path/filepath"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

type commandMatch int

const (
	matchNone commandMatch = iota
	matchFold
	matchExact
)

// matchCommand compares a command name with exe names.
// matchExact means it matches any exe name exactly.
// matchFold means it matches only case-insensitively.
func matchCommand(cmdName string, exeNames []string) commandMatch {
	m := matchNone
	for _, exeName := range exeNames {
		if cmdName == exeName {
			return matchExact
		}
		if strings.EqualFold(cmdName, exeName) {
			m = matchFold
		}
	}
	return m
}

// maybeHasCommand returns true if the package may have a command matching any exe name.
func maybeHasCommand(pkg *aqua.Package, pkgInfo *registry.PackageInfo, exeNames []string) bool {
	for _, exeName := range exeNames {
		if pkgInfo.MaybeHasCommand(exeName) || pkg.HasCommandAlias(exeName) {
			return true
		}
	}
	return false
}

// trimExeExt removes the extension .exe or .bat from the exe name case-insensitively.
// On Windows, a shim can be run with an upper-case extension such as nu.EXE,
// because tools build the command name from PATHEXT.
func trimExeExt(exeName string) string {
	ext := filepath.Ext(exeName)
	if strings.EqualFold(ext, ".exe") || strings.EqualFold(ext, ".bat") {
		return strings.TrimSuffix(exeName, ext)
	}
	return exeName
}
