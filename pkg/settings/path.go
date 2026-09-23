//go:build !windows

package settings

import (
	"path/filepath"

	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

// Path returns the path to the settings file on Unix-like systems.
// The file lives in the XDG configuration directory, so unlike the root
// directory it has no environment variable of its own: a settings file whose
// location depends on the environment would be one more thing to keep in sync.
func Path(osEnv osenv.OSEnv) string {
	xdgConfigHome := osEnv.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(osEnv.Getenv("HOME"), ".config")
	}
	return filepath.Join(xdgConfigHome, "aquaproj-aqua", "config.yaml")
}
