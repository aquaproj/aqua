//go:build windows

package settings

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

// Path returns the path to the settings file on Windows systems.
// It uses the XDG configuration directory with the Windows specific default,
// as GetRootDir does for the data directory.
func Path(osEnv osenv.OSEnv) string {
	xdgConfigHome := xdg.ConfigHome
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(osEnv.Getenv("HOME"), ".config")
	}
	return filepath.Join(xdgConfigHome, "aquaproj-aqua", "config.yaml")
}
