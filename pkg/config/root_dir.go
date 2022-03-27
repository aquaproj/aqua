package config

import (
	"os"
	"path/filepath"
)

type RootDir string

func NewRootDir() RootDir {
	return RootDir(getRootDir())
}

func getRootDir() string {
	if rootDir := os.Getenv("AQUA_ROOT_DIR"); rootDir != "" {
		return rootDir
	}
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(xdgDataHome, "aquaproj-aqua")
}
