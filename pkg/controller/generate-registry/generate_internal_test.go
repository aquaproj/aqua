package genrgst

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestController_getPackageInfo(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		pkgName  string
		exp      *registry.PackageInfo
		releases []*github.RepositoryRelease
		repo     *github.Repository
		assets   []*github.ReleaseAsset
	}{
		{
			name:    "package name doesn't have slash",
			pkgName: "foo",
			exp: &registry.PackageInfo{
				Name: "foo",
				Type: "github_release",
			},
		},
		{
			name:    "repo not found",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner: "foo",
				RepoName:  "foo",
				Type:      "github_release",
			},
		},
		{
			name:    "no release",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner:   "foo",
				RepoName:    "foo",
				Type:        "github_release",
				Description: "hello",
			},
			repo: &github.Repository{
				Description: strP("hello."),
			},
		},
		{
			name:    "normal",
			pkgName: "cli/cli",
			exp: &registry.PackageInfo{
				RepoOwner:   "cli",
				RepoName:    "cli",
				Type:        "github_release",
				Description: "GitHub’s official command line tool",
				Asset:       strP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{Format}}"),
				Format:      "tar.gz",
				Replacements: map[string]string{
					"darwin": "macOS",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "windows",
						Format: "zip",
					},
				},
				SupportedEnvs: []string{
					"darwin",
					"linux",
					"amd64",
				},
				Rosetta2: boolP(true),
			},
			repo: &github.Repository{
				Description: strP("GitHub’s official command line tool"),
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: strP("v2.13.0"),
				},
			},
			assets: []*github.ReleaseAsset{
				{
					Name: strP("gh_2.13.0_linux_amd64.tar.gz"),
				},
				{
					Name: strP("gh_2.13.0_linux_arm64.tar.gz"),
				},
				{
					Name: strP("gh_2.13.0_macOS_amd64.tar.gz"),
				},
				{
					Name: strP("gh_2.13.0_windows_amd64.zip"),
				},
			},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			gh := &github.MockRepositoryService{
				Releases: d.releases,
				Assets:   d.assets,
				Repo:     d.repo,
			}
			ctrl := NewController(nil, gh)
			pkgInfo := ctrl.getPackageInfo(ctx, logE, d.pkgName)
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
