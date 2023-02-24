package util

import (
	"path/filepath"
	"strings"
)

func Ext(s, version string) string {
	return filepath.Ext(strings.ReplaceAll(s, strings.TrimPrefix(version, "v"), ""))
}
