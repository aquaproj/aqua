package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/google/go-cmp/cmp"
)

func TestPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &config.PackageInfo{
				Type: "github_release",
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if name := d.pkgInfo.GetName(); name != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, name)
			}
		})
	}
}

func TestPackageInfo_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp:   "github_release",
			pkgInfo: &config.PackageInfo{
				Type: "github_release",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if typ := d.pkgInfo.GetType(); typ != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, typ)
			}
		})
	}
}

func TestPackageInfo_GetLink(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &config.PackageInfo{
				Type: "github_release",
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if link := d.pkgInfo.GetLink(); link != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, link)
			}
		})
	}
}

func TestPackageInfo_GetDescription(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp:   "hello world",
			pkgInfo: &config.PackageInfo{
				Description: "hello world",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &config.PackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if desc := d.pkgInfo.GetDescription(); desc != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, desc)
			}
		})
	}
}

func TestPackageInfo_GetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &config.PackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &config.PackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if format := d.pkgInfo.GetFormat(); format != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, format)
			}
		})
	}
}

func TestPackageInfo_GetReplacements(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     map[string]string
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp: map[string]string{
				"amd64": "x86_64",
			},
			pkgInfo: &config.PackageInfo{
				Replacements: map[string]string{
					"amd64": "x86_64",
				},
			},
		},
		{
			title:   "empty",
			exp:     nil,
			pkgInfo: &config.PackageInfo{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			replacements := d.pkgInfo.GetReplacements()
			if diff := cmp.Diff(d.exp, replacements); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfo_GetFiles(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     []*config.File
		pkgInfo *config.PackageInfo
	}{
		{
			title: "normal",
			exp: []*config.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &config.PackageInfo{
				Files: []*config.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
			},
		},
		{
			title: "empty",
			exp: []*config.File{
				{
					Name: "ci-info",
				},
			},
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			files := d.pkgInfo.GetFiles()
			if diff := cmp.Diff(d.exp, files); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfo_RenderAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *config.PackageInfo
		pkg     *config.Package
	}{
		{
			title: "github_archive",
			exp:   "",
			pkgInfo: &config.PackageInfo{
				Type: "github_archive",
			},
		},
		{
			title: "github_content",
			exp:   "foo",
			pkgInfo: &config.PackageInfo{
				Type: "github_content",
				Path: template.NewTemplate("foo"),
			},
			pkg: &config.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "github_release",
			exp:   "foo-1.0.0.zip",
			pkgInfo: &config.PackageInfo{
				Type:   "github_release",
				Asset:  template.NewTemplate("foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &config.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "http",
			exp:   "foo-1.0.0.zip",
			pkgInfo: &config.PackageInfo{
				Type:   "http",
				URL:    template.NewTemplate("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &config.Package{
				Version: "v1.0.0",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			asset, err := d.pkgInfo.RenderAsset(d.pkg)
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
		title   string
		exp     string
		pkgInfo *config.PackageInfo
		pkg     *config.Package
	}{
		{
			title: "github_archive",
			exp:   "/tmp/aqua/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkgInfo: &config.PackageInfo{
				Type:      "github_archive",
				RepoOwner: "tfutils",
				RepoName:  "tfenv",
			},
			pkg: &config.Package{
				Version: "v2.2.2",
			},
		},
		{
			title: "github_content",
			exp:   "/tmp/aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v0.2.0/aqua-installer",
			pkgInfo: &config.PackageInfo{
				Type:      "github_content",
				Path:      template.NewTemplate("aqua-installer"),
				RepoOwner: "aquaproj",
				RepoName:  "aqua-installer",
			},
			pkg: &config.Package{
				Version: "v0.2.0",
			},
		},
		{
			title: "github_release",
			exp:   "/tmp/aqua/pkgs/github_release/github.com/aquaproj/aqua/v0.7.7/aqua.tar.gz",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     template.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &config.Package{
				Version: "v0.7.7",
			},
		},
		{
			title: "http",
			exp:   "/tmp/aqua/pkgs/http/example.com/foo-1.0.0.zip",
			pkgInfo: &config.PackageInfo{
				Type:   "http",
				URL:    template.NewTemplate("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &config.Package{
				Version: "v1.0.0",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgPath, err := d.pkgInfo.GetPkgPath(rootDir, d.pkg)
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
		title   string
		exp     string
		pkgInfo *config.PackageInfo
		pkg     *config.Package
		file    *config.File
	}{
		{
			title: "unarchived",
			exp:   "foo",
			pkgInfo: &config.PackageInfo{
				Type: "github_content",
				Path: template.NewTemplate("foo"),
			},
			pkg: &config.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "github_release",
			exp:   "aqua",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     template.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &config.Package{
				Version: "v0.7.7",
			},
			file: &config.File{
				Name: "aqua",
			},
		},
		{
			title: "github_release",
			exp:   "bin/aqua",
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     template.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &config.Package{
				Version: "v0.7.7",
			},
			file: &config.File{
				Name: "aqua",
				Src:  template.NewTemplate("bin/aqua"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			asset, err := d.pkgInfo.GetFileSrc(d.pkg, d.file)
			if err != nil {
				t.Fatal(err)
			}
			if asset != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, asset)
			}
		})
	}
}
