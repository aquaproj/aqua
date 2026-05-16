package config

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func TestPackage_fileSrc(t *testing.T) { //nolint:funlen
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
					Path: "foo",
				},
				Package: &aqua.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: PkgInfoTypeGitHubRelease,
			exp:   pkgNameAqua,
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: repoOwnerAquaproj,
					RepoName:  pkgNameAqua,
					Asset:     "aqua.{{.Format}}",
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: pkgNameAqua,
			},
		},
		{
			title: PkgInfoTypeGitHubRelease,
			exp:   "bin/aqua",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: repoOwnerAquaproj,
					RepoName:  pkgNameAqua,
					Asset:     "aqua.{{.Format}}",
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: pkgNameAqua,
				Src:  "bin/aqua",
			},
		},
		{
			title: "set .exe in windows",
			exp:   "gh.exe",
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: repoOwnerCli,
					RepoName:  repoOwnerCli,
					Asset:     tmplGhAsset,
					Format:    formatZip,
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: "gh",
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
			},
		},
		{
			title: "set .exe in windows (with src)",
			exp:   fileBinGhExe,
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: repoOwnerCli,
					RepoName:  repoOwnerCli,
					Asset:     tmplGhAsset,
					Format:    formatZip,
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: "gh",
				Src:  "bin/gh",
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
			},
		},
		{
			title: "set .exe in windows (src already has .exe)",
			exp:   fileBinGhExe,
			pkg: &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: repoOwnerCli,
					RepoName:  repoOwnerCli,
					Asset:     tmplGhAsset,
					Format:    formatZip,
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: "gh",
				Src:  fileBinGhExe,
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
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
					Path:      "extras/dcgoss/dcgoss",
				},
				Package: &aqua.Package{
					Version: versionV077,
				},
			},
			file: &registry.File{
				Name: "dcgoss",
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
			},
		},
	}
	rt := runtime.New(t.Context())
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if d.rt == nil {
				d.rt = rt
			}
			asset, err := d.pkg.fileSrc(d.file, d.rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}
