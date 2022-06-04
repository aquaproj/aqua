package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
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
				Package: &aqua.Package{
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
				Package: &aqua.Package{
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
				Package: &aqua.Package{
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
			asset, err := d.pkg.RenderAsset(rt)
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
	rootDir := "/tmp/aqua"
	data := []struct {
		title string
		exp   string
		pkg   *config.Package
	}{
		{
			title: "github_archive",
			exp:   "/tmp/aqua/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
				Package: &aqua.Package{
					Version: "v2.2.2",
				},
			},
		},
		{
			title: "github_content",
			exp:   "/tmp/aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v0.2.0/aqua-installer",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					Path:      stringP("aqua-installer"),
					RepoOwner: "aquaproj",
					RepoName:  "aqua-installer",
				},
				Package: &aqua.Package{
					Version: "v0.2.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "/tmp/aqua/pkgs/github_release/github.com/aquaproj/aqua/v0.7.7/aqua.tar.gz",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Asset:     stringP("aqua.{{.Format}}"),
					Format:    "tar.gz",
				},
				Package: &aqua.Package{
					Version: "v0.7.7",
				},
			},
		},
		{
			title: "http",
			exp:   "/tmp/aqua/pkgs/http/example.com/foo-1.0.0.zip",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:   "http",
					URL:    stringP("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
					Format: "zip",
				},
				Package: &aqua.Package{
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
	}{
		{
			title: "unarchived",
			exp:   "foo",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type: "github_content",
					Path: stringP("foo"),
				},
				Package: &aqua.Package{
					Version: "v1.0.0",
				},
			},
		},
		{
			title: "github_release",
			exp:   "aqua",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Asset:     stringP("aqua.{{.Format}}"),
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
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "aquaproj",
					RepoName:  "aqua",
					Asset:     stringP("aqua.{{.Format}}"),
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
	}
	rt := runtime.New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			asset, err := d.pkg.GetFileSrc(d.file, rt)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}
