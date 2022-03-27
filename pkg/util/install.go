package util

import "os"

const OwnerExecutable os.FileMode = 64

func IsOwnerExecutable(mode os.FileMode) bool {
	return mode&OwnerExecutable != 0
}

func AllowOwnerExec(mode os.FileMode) os.FileMode {
	return mode | OwnerExecutable
}
