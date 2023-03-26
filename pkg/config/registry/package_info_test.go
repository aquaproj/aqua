package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func stringP(s string) *string {
	return &s
}

func TestPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
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
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "github_release",
			pkgInfo: &registry.PackageInfo{
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
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
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
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "hello world",
			pkgInfo: &registry.PackageInfo{
				Description: "hello world",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &registry.PackageInfo{},
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
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &registry.PackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &registry.PackageInfo{},
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
		exp     registry.Replacements
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp: registry.Replacements{
				"amd64": "x86_64",
			},
			pkgInfo: &registry.PackageInfo{
				Replacements: registry.Replacements{
					"amd64": "x86_64",
				},
			},
		},
		{
			title:   "empty",
			exp:     nil,
			pkgInfo: &registry.PackageInfo{},
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

func TestPackageInfo_GetFiles(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		exp     []*registry.File
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp: []*registry.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Files: []*registry.File{
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
			exp: []*registry.File{
				{
					Name: "ci-info",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "has name",
			exp: []*registry.File{
				{
					Name: "cmctl",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "cert-manager",
				RepoName:  "cert-manager",
				Name:      "cert-manager/cert-manager/cmctl",
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

func TestPackageInfo_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title:   "package name is required",
			pkgInfo: &registry.PackageInfo{},
			isErr:   true,
		},
		{
			title: "repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubArchive,
			},
			isErr: true,
		},
		{
			title: "github_archive",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubArchive,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "github_content repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: "github_content path is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_content",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Path:      stringP("bin/ci-info"),
			},
		},
		{
			title: "github_release repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubRelease,
			},
			isErr: true,
		},
		{
			title: "github_release asset is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info.tar.gz"),
			},
		},
		{
			title: "http url is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
			},
			isErr: true,
		},
		{
			title: "http",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
				Name: "suzuki-shunsuke/ci-info",
				URL:  stringP("http://example.com"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
