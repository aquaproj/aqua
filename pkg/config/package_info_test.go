package config_test

import (
	"testing"

	"github.com/clivm/clivm/pkg/config"
	"github.com/clivm/clivm/pkg/config/clivm"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/runtime"
)

func stringP(s string) *string {
	return &s
}

func TestPackage_RenderAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
		rt    *runtime.Runtime
	}{
		{
			title: "github_archive",
			exp:   "",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: "github_archive",
				},
			},
		},
		{
			title: "github_content",
			exp:   "foo",
			pkg: &config.Package{
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type: "github_content",
					Path: stringP("foo"),
				},
			},
		},
		{
			title: "github_release",
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "github_release",
					Asset:  stringP("foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "http",
			exp:   "foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "http",
					URL:    stringP("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "windows add .exe",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "github_release",
					Asset:  stringP("foo-{{.OS}}-{{.Arch}}"),
					Format: "raw",
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "windows add .exe without Format",
			exp:   "foo-windows-amd64.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  "github_release",
					Asset: stringP("foo-{{.OS}}-{{.Arch}}"),
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
		},
		{
			title: "windows",
			exp:   "foo-windows-amd64.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:  "github_release",
					Asset: stringP("foo-{{.OS}}-{{.Arch}}.tar.gz"),
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
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

func TestPackageInfo_GetPkgPath(t *testing.T) { //nolint:funlen
	t.Parallel()
	rootDir := "/tmp/clivm"
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
	}{
		{
			title: "github_archive",
			exp:   "/tmp/clivm/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
				Package: &clivm.Package{
					Version: "v2.2.2",
				},
			},
		},
		{
			title: "github_content",
			exp:   "/tmp/clivm/pkgs/github_content/github.com/clivm/clivm-installer/v0.2.0/clivm-installer",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					Path:      stringP("clivm-installer"),
					RepoOwner: "clivm",
					RepoName:  "clivm-installer",
				},
				Package: &clivm.Package{
					Version: "v0.2.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "/tmp/clivm/pkgs/github_release/github.com/clivm/clivm/v0.7.7/clivm.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "clivm",
					RepoName:  "clivm",
					Asset:     stringP("clivm.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &clivm.Package{
					Version: "v0.7.7",
				},
			},
		},
		{
			title: "http",
			exp:   "/tmp/clivm/pkgs/http/example.com/foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "http",
					URL:    stringP("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
		},
	}
	rt := runtime.New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgPath, err := d.pkg.GetPkgPath(rootDir, rt)
			if err != nil {
				t.Fatal(err)
			}
			if pkgPath != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, pkgPath)
			}
		})
	}
}

func TestPackageInfo_GetFileSrc(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
		file  *registry.File
		rt    *runtime.Runtime
	}{
		{
			title: "unarchived",
			exp:   "foo",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: "github_content",
					Path: stringP("foo"),
				},
				Package: &clivm.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "clivm",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "clivm",
					RepoName:  "clivm",
					Asset:     stringP("clivm.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &clivm.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "clivm",
			},
		},
		{
			title: "github_release",
			exp:   "bin/clivm",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "clivm",
					RepoName:  "clivm",
					Asset:     stringP("clivm.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &clivm.Package{
					Version: "v0.7.7",
				},
			},
			file: &registry.File{
				Name: "clivm",
				Src:  "bin/clivm",
			},
		},
		{
			title: "set .exe in windows",
			exp:   "gh.exe",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     stringP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &clivm.Package{
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
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     stringP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &clivm.Package{
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
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     stringP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "zip",
				},
				Package: &clivm.Package{
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
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Name:      "aelsabbahy/goss/dcgoss",
					Type:      "github_content",
					RepoOwner: "aelsabbahy",
					RepoName:  "goss",
					Path:      stringP("extras/dcgoss/dcgoss"),
				},
				Package: &clivm.Package{
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
			asset, err := d.pkg.GetFileSrc(d.file, d.rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}
