//go:build windows

package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

// GetRootDir determines the root directory for aqua installation on Windows systems.
// It checks AQUA_ROOT_DIR environment variable first, then uses XDG data directory with Windows-specific defaults.
func GetRootDir(osEnv osenv.OSEnv) string {
	if rootDir := osEnv.Getenv("AQUA_ROOT_DIR"); rootDir != "" {
		return rootDir
	}
	xdgDataHome := xdg.DataHome
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(osEnv.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(xdgDataHome, "aquaproj-aqua")
}
