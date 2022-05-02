package config

import (
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
)

type OverrideSimple struct {
	GOOS         string            `json:"goos,omitempty" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=js,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	GOArch       string            `json:"goarch,omitempty" jsonschema:"enum=386,enum=amd64,enum=arm,enum=arm64,enum=mips,enum=mips64,enum=mips64le,enum=mipsle,enum=ppc64,enum=ppc64le,enum=riscv64,enum=s390x,enum=wasm"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Format       string            `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Asset        string            `json:"asset,omitempty"`
	Files        []*FileSimple     `json:"files,omitempty"`
	URL          string            `json:"url,omitempty"`
}

type Override struct {
	GOOS         string             `json:"goos,omitempty" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=js,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	GOArch       string             `json:"goarch,omitempty" jsonschema:"enum=386,enum=amd64,enum=arm,enum=arm64,enum=mips,enum=mips64,enum=mips64le,enum=mipsle,enum=ppc64,enum=ppc64le,enum=riscv64,enum=s390x,enum=wasm"`
	Replacements map[string]string  `json:"replacements,omitempty"`
	Format       string             `json:"format,omitempty" jsonschema:"example=tar.gz,example=raw"`
	Asset        *template.Template `json:"asset,omitempty"`
	Files        []*File            `json:"files,omitempty"`
	URL          *template.Template `json:"url,omitempty"`
}

func (ov *Override) Match(rt *runtime.Runtime) bool {
	if ov.GOOS != "" && ov.GOOS != rt.GOOS {
		return false
	}
	if ov.GOArch != "" && ov.GOArch != rt.GOARCH {
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
