package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func TestPackage_RenderAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
		rt    *runtime.Runtime
	}{
		{
			title: pkgTypeGitHubArchive,
			exp:   "",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: pkgTypeGitHubArchive,
				},
			},
		},
		{
			title: pkgTypeGitHubContent,
			exp:   "foo",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: versionV1,
				},
				PackageInfo: &registry.PackageInfo{
					Type: pkgTypeGitHubContent,
					Path: "foo",
				},
			},
		},
		{
			title: pkgTypeGitHubRelease,
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   pkgTypeGitHubRelease,
					Asset:  "foo-{{trimV .Version}}.{{.Format}}",
					Format: formatZip,
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
			},
		},
		{
			title: pkgTypeHTTP,
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   pkgTypeHTTP,
					URL:    "https://example.com/foo-{{trimV .Version}}.{{.Format}}",
					Format: formatZip,
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
			},
		},
		{
			title: "windows add .exe",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   pkgTypeGitHubRelease,
					Asset:  "foo-{{.OS}}-{{.Arch}}",
					Format: "raw",
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
			},
		},
		{
			title: "windows add .exe without Format",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  pkgTypeGitHubRelease,
					Asset: "foo-{{.OS}}-{{.Arch}}",
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
			},
			rt: &runtime.Runtime{
				GOOS:   osWindows,
				GOARCH: archAmd64,
			},
		},
		{
			title: osWindows,
			exp:   "foo-windows-amd64.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  pkgTypeGitHubRelease,
					Asset: "foo-{{.OS}}-{{.Arch}}.tar.gz",
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
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
			asset, err := d.pkg.RenderAsset(d.rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}

func TestPackageInfo_PkgPath(t *testing.T) { //nolint:funlen
	t.Parallel()
	rootDir := "/tmp/aqua"
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
	}{
		{
			title: pkgTypeGitHubArchive,
			exp:   "/tmp/aqua/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      pkgTypeGitHubArchive,
					RepoOwner: repoOwnerTfutils,
					RepoName:  repoNameTfenv,
				},
				Package: &aqua.Package{
					Version: "v2.2.2",
				},
			},
		},
		{
			title: pkgTypeGitHubContent,
			exp:   "/tmp/aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v0.2.0/aqua-installer",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      pkgTypeGitHubContent,
					Path:      repoNameAquaInstaller,
					RepoOwner: repoOwnerAquaproj,
					RepoName:  repoNameAquaInstaller,
				},
				Package: &aqua.Package{
					Version: "v0.2.0",
				},
			},
		},
		{
			title: pkgTypeGitHubRelease,
			exp:   "/tmp/aqua/pkgs/github_release/github.com/suzuki-shunsuke/ci-info/v0.7.7/ci-info.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      pkgTypeGitHubRelease,
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Asset:     "ci-info.{{.Format}}",
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
		},
		{
			title: pkgTypeHTTP,
			exp:   "/tmp/aqua/pkgs/http/example.com/foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   pkgTypeHTTP,
					URL:    "https://example.com/foo-{{trimV .Version}}.{{.Format}}",
					Format: formatZip,
				},
				Package: &aqua.Package{
					Version: versionV1,
				},
			},
		},
	}
	rt := runtime.New(t.Context())
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgPath, err := d.pkg.AbsPkgPath(rootDir, rt)
			if err != nil {
				t.Fatal(err)
			}
			if pkgPath != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, pkgPath)
			}
		})
	}
}
