package osfile

import "os"

const ownerExecutable os.FileMode = 64

func IsOwnerExecutable(mode os.FileMode) bool {
	return mode&ownerExecutable != 0
}

// IsExecutable reports whether a file with the given mode can be executed on the given OS.
// Windows file modes don't have executable bits, so any regular file is treated as executable on Windows.
func IsExecutable(goos string, mode os.FileMode) bool {
	if goos == "windows" {
		return mode.IsRegular()
	}
	return IsOwnerExecutable(mode)
}

func AllowOwnerExec(mode os.FileMode) os.FileMode {
	return mode | ownerExecutable
}
