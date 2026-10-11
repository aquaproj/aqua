package packslip

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

const pkgName = "packslip"

func Package() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "jdx/packslip",
			Version: Version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:               "github_release",
			RepoOwner:          "jdx",
			RepoName:           pkgName,
			VersionConstraints: `false`,
			Asset:              "packslip-{{.Version}}-{{.OS}}-{{.Arch}}.{{.Format}}",
			Format:             "tar.xz",
			Files: []*registry.File{
				{
					Name: pkgName,
					Src:  "{{.AssetWithoutExt}}/packslip",
				},
			},
			Replacements: map[string]string{
				"amd64": "x64",
			},
			Overrides: []*registry.Override{
				{
					GOOS:   "windows",
					Format: "zip",
				},
			},
			SupportedEnvs: []string{
				"linux",
				"darwin/arm64",
				"windows",
			},
		},
	}
}
