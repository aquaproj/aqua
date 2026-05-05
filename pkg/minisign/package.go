package minisign

import (
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func Package() *config.Package { //nolint:funlen
	pkg := &aqua.Package{
		Name:    "jedisct1/minisign",
		Version: Version,
	}
	if Version == "0.11" {
		return &config.Package{
			Package: pkg,
			PackageInfo: &registry.PackageInfo{
				Type:                "github_release",
				RepoOwner:           "jedisct1",
				RepoName:            pkgName,
				VersionConstraints:  `false`,
				Asset:               "minisign-{{.Version}}-{{.OS}}.{{.Format}}",
				Format:              "zip",
				Rosetta2:            true,
				WindowsARMEmulation: true,
				Replacements: map[string]string{
					osDarwin:  "macos",
					osWindows: "win64",
					archAmd64: "x86_64",
					"arm64":   "aarch64",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "linux",
						Format: "tar.gz",
						Files: []*registry.File{
							{
								Name: pkgName,
								Src:  "minisign-{{.OS}}/{{.Arch}}/minisign",
							},
						},
					},
					{
						GOOS: osWindows,
						Files: []*registry.File{
							{
								Name: pkgName,
								Src:  "minisign-win64/minisign.exe",
							},
						},
					},
				},
				SupportedEnvs: []string{
					osDarwin,
					osWindows,
					archAmd64,
				},
			},
		}
	}
	return &config.Package{
		Package: pkg,
		PackageInfo: &registry.PackageInfo{
			Type:                "github_release",
			RepoOwner:           "jedisct1",
			RepoName:            pkgName,
			VersionConstraints:  `false`,
			Asset:               "minisign-{{.Version}}-{{.OS}}.{{.Format}}",
			Format:              "zip",
			WindowsARMEmulation: true,
			Files: []*registry.File{
				{
					Name: pkgName,
					Src:  "minisign-{{.OS}}/{{.Arch}}/minisign",
				},
			},
			Replacements: map[string]string{
				osDarwin:  "macos",
				osWindows: "win64",
				archAmd64: "x86_64",
				"arm64":   "aarch64",
			},
			Overrides: []*registry.Override{
				{
					GOOS:   "linux",
					Format: "tar.gz",
				},
				{
					GOOS: osDarwin,
					Files: []*registry.File{
						{
							Name: pkgName,
						},
					},
				},
			},
			SupportedEnvs: []string{
				"darwin/arm64",
				osWindows,
				"linux/amd64",
			},
		},
	}
}
