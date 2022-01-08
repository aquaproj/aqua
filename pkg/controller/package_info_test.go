package controller_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &controller.PackageInfo{
				Type: "github_release",
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &controller.PackageInfo{
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
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp:   "github_release",
			pkgInfo: &controller.PackageInfo{
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
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &controller.PackageInfo{
				Type: "github_release",
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &controller.PackageInfo{
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
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp:   "hello world",
			pkgInfo: &controller.PackageInfo{
				Description: "hello world",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.PackageInfo{},
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
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &controller.PackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.PackageInfo{},
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
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp: map[string]string{
				"amd64": "x86_64",
			},
			pkgInfo: &controller.PackageInfo{
				Replacements: map[string]string{
					"amd64": "x86_64",
				},
			},
		},
		{
			title:   "empty",
			exp:     nil,
			pkgInfo: &controller.PackageInfo{},
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
		exp     []*controller.File
		pkgInfo *controller.PackageInfo
	}{
		{
			title: "normal",
			exp: []*controller.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &controller.PackageInfo{
				Files: []*controller.File{
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
			exp: []*controller.File{
				{
					Name: "ci-info",
				},
			},
			pkgInfo: &controller.PackageInfo{
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
		pkgInfo *controller.PackageInfo
		pkg     *controller.Package
	}{
		{
			title: "github_archive",
			exp:   "",
			pkgInfo: &controller.PackageInfo{
				Type: "github_archive",
			},
		},
		{
			title: "github_content",
			exp:   "foo",
			pkgInfo: &controller.PackageInfo{
				Type: "github_content",
				Path: controller.NewTemplate("foo"),
			},
			pkg: &controller.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "github_release",
			exp:   "foo-1.0.0.zip",
			pkgInfo: &controller.PackageInfo{
				Type:   "github_release",
				Asset:  controller.NewTemplate("foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &controller.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "http",
			exp:   "foo-1.0.0.zip",
			pkgInfo: &controller.PackageInfo{
				Type:   "http",
				URL:    controller.NewTemplate("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &controller.Package{
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
		pkgInfo *controller.PackageInfo
		pkg     *controller.Package
	}{
		{
			title: "github_archive",
			exp:   "/tmp/aqua/pkgs/github_archive/github.com/tfutils/tfenv/v2.2.2",
			pkgInfo: &controller.PackageInfo{
				Type:      "github_archive",
				RepoOwner: "tfutils",
				RepoName:  "tfenv",
			},
			pkg: &controller.Package{
				Version: "v2.2.2",
			},
		},
		{
			title: "github_content",
			exp:   "/tmp/aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v0.2.0/aqua-installer",
			pkgInfo: &controller.PackageInfo{
				Type:      "github_content",
				Path:      controller.NewTemplate("aqua-installer"),
				RepoOwner: "aquaproj",
				RepoName:  "aqua-installer",
			},
			pkg: &controller.Package{
				Version: "v0.2.0",
			},
		},
		{
			title: "github_release",
			exp:   "/tmp/aqua/pkgs/github_release/github.com/aquaproj/aqua/v0.7.7/aqua.tar.gz",
			pkgInfo: &controller.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     controller.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &controller.Package{
				Version: "v0.7.7",
			},
		},
		{
			title: "http",
			exp:   "/tmp/aqua/pkgs/http/example.com/foo-1.0.0.zip",
			pkgInfo: &controller.PackageInfo{
				Type:   "http",
				URL:    controller.NewTemplate("https://example.com/foo-{{trimV .Version}}.{{.Format}}"),
				Format: "zip",
			},
			pkg: &controller.Package{
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
		pkgInfo *controller.PackageInfo
		pkg     *controller.Package
		file    *controller.File
	}{
		{
			title: "unarchived",
			exp:   "foo",
			pkgInfo: &controller.PackageInfo{
				Type: "github_content",
				Path: controller.NewTemplate("foo"),
			},
			pkg: &controller.Package{
				Version: "v1.0.0",
			},
		},
		{
			title: "github_release",
			exp:   "aqua",
			pkgInfo: &controller.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     controller.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &controller.Package{
				Version: "v0.7.7",
			},
			file: &controller.File{
				Name: "aqua",
			},
		},
		{
			title: "github_release",
			exp:   "bin/aqua",
			pkgInfo: &controller.PackageInfo{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     controller.NewTemplate("aqua.{{.Format}}"),
				Format:    "tar.gz",
			},
			pkg: &controller.Package{
				Version: "v0.7.7",
			},
			file: &controller.File{
				Name: "aqua",
				Src:  controller.NewTemplate("bin/aqua"),
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

func TestPackageInfo_SetVersion(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		version string
		pkgInfo *controller.PackageInfo
		exp     *controller.PackageInfo
	}{
		{
			title: "no version constraint",
			exp: &controller.PackageInfo{
				Type: "github_content",
				Path: controller.NewTemplate("foo"),
			},
			pkgInfo: &controller.PackageInfo{
				Type: "github_content",
				Path: controller.NewTemplate("foo"),
			},
		},
		{
			title: "version constraint",
			exp: &controller.PackageInfo{
				Type:               "github_content",
				Path:               controller.NewTemplate("foo"),
				VersionConstraints: controller.NewVersionConstraints(`semver(">= 0.4.0")`),
			},
			pkgInfo: &controller.PackageInfo{
				Type:               "github_content",
				Path:               controller.NewTemplate("foo"),
				VersionConstraints: controller.NewVersionConstraints(`semver(">= 0.4.0")`),
			},
			version: "v0.5.0",
		},
		{
			title: "child version constraint",
			exp: &controller.PackageInfo{
				Type:               "github_content",
				Path:               controller.NewTemplate("bar"),
				VersionConstraints: controller.NewVersionConstraints(`semver(">= 0.4.0")`),
				VersionOverrides: []*controller.PackageInfo{
					{
						VersionConstraints: controller.NewVersionConstraints(`semver("< 0.4.0")`),
						Path:               controller.NewTemplate("bar"),
					},
				},
			},
			pkgInfo: &controller.PackageInfo{
				Type:               "github_content",
				Path:               controller.NewTemplate("foo"),
				VersionConstraints: controller.NewVersionConstraints(`semver(">= 0.4.0")`),
				VersionOverrides: []*controller.PackageInfo{
					{
						VersionConstraints: controller.NewVersionConstraints(`semver("< 0.4.0")`),
						Path:               controller.NewTemplate("bar"),
					},
				},
			},
			version: "v0.3.0",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.SetVersion(d.version)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(pkgInfo, d.exp, cmpopts.IgnoreUnexported(controller.VersionConstraints{}), cmp.AllowUnexported(controller.Template{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
