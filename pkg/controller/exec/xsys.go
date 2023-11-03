package exec

import (
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func getEnabledXSysExec(osEnv osenv.OSEnv, goos string) bool {
	if goos == "windows" {
		return false
	}
	if osEnv.Getenv("AQUA_EXPERIMENTAL_X_SYS_EXEC") == "false" {
		return false
	}
	if osEnv.Getenv("AQUA_X_SYS_EXEC") == "false" {
		return false
	}
	return true
}
