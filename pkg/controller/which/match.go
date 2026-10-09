package which

import (
	"path/filepath"
	"strings"
)

type commandMatch int

const (
	matchNone commandMatch = iota
	matchFold
	matchExact
)

// matchCommand compares a command name with an exe name.
// matchFold means they match only case-insensitively.
func matchCommand(cmdName, exeName string) commandMatch {
	if cmdName == exeName {
		return matchExact
	}
	if strings.EqualFold(cmdName, exeName) {
		return matchFold
	}
	return matchNone
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
