package config

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/util"
)

func TestPackage_getFileSrc(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		exp   string
		pkg   *Package
		file  *registry.File
		rt    *runtime.Runtime
	}{
		{
			title: "unarchived",
			exp:   "foo",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type: "github_content",
					Path: util.StrP("foo"),
				},
				Package: &aqua.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "aqua",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Asset:     util.StrP("aqua.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "aqua",
			},
		},
		{
			title: "github_release",
			exp:   "bin/aqua",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Asset:     util.StrP("aqua.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "aqua",
				Src:  "bin/aqua",
			},
		},
		{
			title: "set .exe in windows",
			exp:   "gh.exe",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     util.StrP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "gh",
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "set .exe in windows (with src)",
			exp:   "bin/gh.exe",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     util.StrP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "gh",
				Src:  "bin/gh",
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "set .exe in windows (src already has .exe)",
			exp:   "bin/gh.exe",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     util.StrP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "gh",
				Src:  "bin/gh.exe",
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "add .sh in case of github_content",
			exp:   "dcgoss.sh",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Name:      "aelsabbahy/goss/dcgoss",
					Type:      "github_content",
					RepoOwner: "aelsabbahy",
					RepoName:  "goss",
					Path:      util.StrP("extras/dcgoss/dcgoss"),
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "dcgoss",
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
	}
	rt := runtime.New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if d.rt == nil {
				d.rt = rt
			}
			asset, err := d.pkg.getFileSrc(d.file, d.rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}
