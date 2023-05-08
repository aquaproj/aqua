package registry

import (
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type Override struct {
	GOOS               string          `yaml:",omitempty" json:"goos,omitempty" jsonschema:"enum=aix,enum=android,enum=darwin,enum=dragonfly,enum=freebsd,enum=illumos,enum=ios,enum=linux,enum=netbsd,enum=openbsd,enum=plan9,enum=solaris,enum=windows"`
	GOArch             string          `yaml:",omitempty" json:"goarch,omitempty" jsonschema:"enum=386,enum=amd64,enum=arm,enum=arm64,enum=mips,enum=mips64,enum=mips64le,enum=mipsle,enum=ppc64,enum=ppc64le,enum=riscv64,enum=s390x"`
	Type               string          `json:"type,omitempty" jsonschema:"enum=github_release,enum=github_content,enum=github_archive,enum=http,enum=go,enum=go_install"`
	Format             string          `yaml:",omitempty" json:"format,omitempty" jsonschema:"example=tar.gz,example=raw,example=zip"`
	Asset              *string         `yaml:",omitempty" json:"asset,omitempty"`
	Files              []*File         `yaml:",omitempty" json:"files,omitempty"`
	URL                *string         `yaml:",omitempty" json:"url,omitempty"`
	CompleteWindowsExt *bool           `json:"complete_windows_ext,omitempty" yaml:"complete_windows_ext,omitempty"`
	WindowsExt         string          `json:"windows_ext,omitempty" yaml:"windows_ext,omitempty"`
	Replacements       Replacements    `yaml:",omitempty" json:"replacements,omitempty"`
	Checksum           *Checksum       `json:"checksum,omitempty"`
	Cosign             *Cosign         `json:"cosign,omitempty"`
	SLSAProvenance     *SLSAProvenance `json:"slsa_provenance,omitempty" yaml:"slsa_provenance,omitempty"`
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
