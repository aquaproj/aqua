package genrgst

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestController_getPackageInfo(t *testing.T) { //nolint:funlen
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
				Description: util.StrP("hello."),
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
				Asset:       util.StrP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
				Format:      "tar.gz",
				Replacements: registry.Replacements{
					"darwin": "macOS",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "windows",
						Format: "zip",
					},
				},
				SupportedEnvs: registry.SupportedEnvs{
					"darwin",
					"linux",
					"amd64",
				},
				Rosetta2: util.BoolP(true),
			},
			repo: &github.Repository{
				Description: util.StrP("GitHub’s official command line tool"),
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: util.StrP("v2.13.0"),
				},
			},
			assets: []*github.ReleaseAsset{
				{
					Name: util.StrP("gh_2.13.0_linux_amd64.tar.gz"),
				},
				{
					Name: util.StrP("gh_2.13.0_linux_arm64.tar.gz"),
				},
				{
					Name: util.StrP("gh_2.13.0_macOS_amd64.tar.gz"),
				},
				{
					Name: util.StrP("gh_2.13.0_windows_amd64.zip"),
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
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
				Assets:   d.assets,
				Repo:     d.repo,
			}
			ctrl := NewController(nil, gh, nil)
			pkgInfo, _ := ctrl.getPackageInfo(ctx, logE, d.pkgName, true, nil)
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
