package config

import (
	"runtime"

	"github.com/aquaproj/aqua/pkg/template"
)

type Override struct {
	GOOS         string
	GOArch       string
	Replacements map[string]string
	Format       string
	Asset        *template.Template
	Files        []*File
	URL          *template.Template
}

func (ov *Override) Match() bool {
	if ov.GOOS != "" && ov.GOOS != runtime.GOOS {
		return false
	}
	if ov.GOArch != "" && ov.GOArch != runtime.GOARCH {
		return false
	}
	return true
}

func replace(key string, replacements map[string]string) string {
	a := replacements[key]
	if a == "" {
		return key
	}
	return a
}
