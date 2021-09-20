package controller_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/suzuki-shunsuke/aqua/pkg/controller"
)

func TestGitHubReleasePackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &controller.GitHubReleasePackageInfo{
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &controller.GitHubReleasePackageInfo{
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

func TestGitHubReleasePackageInfo_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title:   "normal",
			exp:     "github_release",
			pkgInfo: &controller.GitHubReleasePackageInfo{},
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

func TestGitHubReleasePackageInfo_GetLink(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &controller.GitHubReleasePackageInfo{
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &controller.GitHubReleasePackageInfo{
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

func TestGitHubReleasePackageInfo_GetDescription(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title: "normal",
			exp:   "hello world",
			pkgInfo: &controller.GitHubReleasePackageInfo{
				Description: "hello world",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.GitHubReleasePackageInfo{},
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

func TestGitHubReleasePackageInfo_GetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &controller.GitHubReleasePackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &controller.GitHubReleasePackageInfo{},
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

func TestGitHubReleasePackageInfo_GetReplacements(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     map[string]string
		pkgInfo *controller.GitHubReleasePackageInfo
	}{
		{
			title: "normal",
			exp: map[string]string{
				"amd64": "x86_64",
			},
			pkgInfo: &controller.GitHubReleasePackageInfo{
				Replacements: map[string]string{
					"amd64": "x86_64",
				},
			},
		},
		{
			title:   "empty",
			exp:     nil,
			pkgInfo: &controller.GitHubReleasePackageInfo{},
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

func TestGitHubReleasePackageInfo_GetFiles(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     []*controller.File
		pkgInfo *controller.GitHubReleasePackageInfo
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
			pkgInfo: &controller.GitHubReleasePackageInfo{
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
			pkgInfo: &controller.GitHubReleasePackageInfo{
				RepoName: "ci-info",
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
