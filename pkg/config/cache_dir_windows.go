//go:build windows

package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

// GetCacheDir returns the directory aqua caches downloads in.
//
// It is the XDG cache directory rather than the root directory, because what goes
// here can be thrown away and rebuilt, while the root directory holds the installed
// packages themselves.
func GetCacheDir(osEnv osenv.OSEnv) string {
	xdgCacheHome := xdg.CacheHome
	if xdgCacheHome == "" {
		xdgCacheHome = filepath.Join(osEnv.Getenv("HOME"), ".cache")
	}
	return filepath.Join(xdgCacheHome, "aquaproj-aqua")
}
