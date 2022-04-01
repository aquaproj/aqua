package config

import (
	"runtime"

	"github.com/aquaproj/aqua/pkg/template"
)

type Override struct {
	GOOS         string             `json:"goos"`
	GOArch       string             `json:"goarch"`
	Replacements map[string]string  `json:"replacements"`
	Format       string             `json:"format"`
	Asset        *template.Template `json:"asset"`
	Files        []*File            `json:"files"`
	URL          *template.Template `json:"url"`
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
