package util

import "os"

const ownerExecutable os.FileMode = 64

func IsOwnerExecutable(mode os.FileMode) bool {
	return mode&ownerExecutable != 0
}

func AllowOwnerExec(mode os.FileMode) os.FileMode {
	return mode | ownerExecutable
}
