package config

import (
	"runtime"

	"github.com/aquaproj/aqua/pkg/template"
)

type Override struct {
	GOOS         string             `json:"goos,omitempty" jsonschema:"example=darwin,example=linux"`
	GOArch       string             `json:"goarch,omitempty" jsonschema:"example=amd64,example=arm64"`
	Replacements map[string]string  `json:"replacements,omitempty"`
	Format       string             `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Asset        *template.Template `json:"asset,omitempty"`
	Files        []*File            `json:"files,omitempty"`
	URL          *template.Template `json:"url,omitempty"`
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
