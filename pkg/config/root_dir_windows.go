//go:build windows
// +build windows

package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func GetRootDir(osEnv osenv.OSEnv) string {
	if rootDir := osEnv.Getenv("CLIVM_ROOT_DIR"); rootDir != "" {
		return rootDir
	}
	xdgDataHome := xdg.DataHome
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(osEnv.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(xdg.DataHome, "clivm")
}
