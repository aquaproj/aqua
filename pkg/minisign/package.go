package minisign

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
)

func Package() *config.Package { //nolint:funlen
	return &config.Package{
		Package: &aqua.Package{
			Name:    "jedisct1/minisign",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "jedisct1",
			RepoName:  "minisign",
			VersionOverrides: []*registry.VersionOverride{
				{
					VersionConstraints:  `Version == "0.11"`,
					Asset:               "minisign-{{.Version}}-{{.OS}}.{{.Format}}",
					Format:              "zip",
					Rosetta2:            ptr.Bool(true),
					WindowsARMEmulation: ptr.Bool(true),
					Replacements: map[string]string{
						"darwin":  "macos",
						"windows": "win64",
						"amd64":   "x86_64",
						"arm64":   "aarch64",
					},
					Overrides: []*registry.Override{
						{
							GOOS:   "linux",
							Format: "tar.gz",
							Files: []*registry.File{
								{
									Name: "minisign",
									Src:  "minisign-{{.OS}}/{{.Arch}}/minisign",
								},
							},
						},
						{
							GOOS: "windows",
							Files: []*registry.File{
								{
									Name: "minisign",
									Src:  "minisign-win64/minisign.exe",
								},
							},
						},
					},
					SupportedEnvs: []string{
						"darwin",
						"windows",
						"amd64",
					},
				},
				{
					VersionConstraints:  `true`,
					Asset:               "minisign-{{.Version}}-{{.OS}}.{{.Format}}",
					Format:              "zip",
					WindowsARMEmulation: ptr.Bool(true),
					Files: []*registry.File{
						{
							Name: "minisign",
							Src:  "minisign-{{.OS}}/{{.Arch}}/minisign",
						},
					},
					Replacements: map[string]string{
						"darwin":  "macos",
						"windows": "win64",
						"amd64":   "x86_64",
						"arm64":   "aarch64",
					},
					Overrides: []*registry.Override{
						{
							GOOS:   "linux",
							Format: "tar.gz",
						},
						{
							GOOS: "darwin",
							Files: []*registry.File{
								{
									Name: "minisign",
								},
							},
						},
					},
					SupportedEnvs: []string{
						"darwin/arm64",
						"windows",
						"linux/amd64",
					},
				},
			},
		},
	}
}
