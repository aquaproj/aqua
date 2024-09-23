package ghattestation

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func Package() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "cli/cli",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "cli",
			RepoName:  "cli",
			Asset:     "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
			Format:    "zip",
			Files: []*registry.File{
				{
					Name: "gh",
					Src:  "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}/bin/gh",
				},
			},
			Replacements: map[string]string{
				"darwin": "macOS",
			},
			Overrides: []*registry.Override{
				{
					GOOS:   "linux",
					Format: "tar.gz",
				},
				{
					GOOS: "windows",
					Files: []*registry.File{
						{
							Name: "gh",
							Src:  "bin/gh.exe",
						},
					},
				},
			},
		},
	}
}
